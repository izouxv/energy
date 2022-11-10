package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"encoding/json"
	"errors"
	"fmt"
	progressbar "github.com/energye/energy/cmd/progress-bar"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var CmdInstall = &Command{
	UsageLine: "install [path] [version]",
	Short:     "Automatically configure the CEF and Energy framework",
	Long: `Automatically configure the CEF and Energy framework
During this process, CEF and Energy are downloaded
Default framework name is "CEFEnergy"`,
}

const (
	cefKey                      = "cef"
	energyKey                   = "energy"
	download_version_config_url = "https://energy.yanghy.cn/autoconfig/%s.json"
	download_extract_url        = "https://energy.yanghy.cn/autoconfig/extract.json"
	frameworkCache              = "frameworkCache"
)

type downloadInfo struct {
	fileName      string
	frameworkPath string
	downloadPath  string
	url           string
	success       bool
}

func init() {
	CmdInstall.Run = runInstall
}

//运行安装
func runInstall(c *CommandConfig) error {
	if c.Install.Path == "" {
		c.Install.Path = c.Wd
	}
	c.Install.Path = filepath.Join(c.Install.Path, c.Install.Name)
	if c.Install.Version == "" {
		c.Install.Version = "latest"
	}
	os.MkdirAll(c.Install.Path, fs.ModeDir)
	os.MkdirAll(filepath.Join(c.Install.Path, frameworkCache), fs.ModeDir)
	println("Start downloading CEF and Energy dependency")
	downloadJSON, err := downloadConfig(fmt.Sprintf(download_version_config_url, c.Install.Version))
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	var downloadVersion map[string]interface{}
	downloadJSON = bytes.TrimPrefix(downloadJSON, []byte("\xef\xbb\xbf"))
	if err := json.Unmarshal(downloadJSON, &downloadVersion); err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	version, ok := downloadVersion[c.Install.Version]
	if !ok {
		println("Invalid version number:", c.Install.Version)
		os.Exit(1)
	}
	osConfig := version.(map[string]interface{})[runtime.GOOS]
	var osVersion map[string]interface{}
	if runtime.GOOS == "windows" {
		//区分windows系统位数
		bits := osConfig.(map[string]interface{})
		osVersion = bits[fmt.Sprintf("%d", strconv.IntSize)].(map[string]interface{})
	}
	//提取文件配置
	extractData, err := downloadConfig(download_extract_url)
	if err != nil {
		fmt.Errorf("%s", err.Error())
		os.Exit(1)
	}
	var extractConfig map[string]interface{}
	extractData = bytes.TrimPrefix(extractData, []byte("\xef\xbb\xbf"))
	if err := json.Unmarshal(extractData, &extractConfig); err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	extractOSConfig := extractConfig[runtime.GOOS].(map[string]interface{})

	//下载地址
	var (
		cefUrl, cefOk       = osVersion[cefKey].(string)
		energyUrl, energyOk = osVersion[energyKey].(string)
	)
	if cefOk && energyOk {
		var downloads = make(map[string]*downloadInfo)
		downloads[cefKey] = &downloadInfo{fileName: urlName(cefUrl), downloadPath: filepath.Join(c.Install.Path, frameworkCache, urlName(cefUrl)), frameworkPath: c.Install.Path, url: cefUrl}
		downloads[energyKey] = &downloadInfo{fileName: urlName(energyUrl), downloadPath: filepath.Join(c.Install.Path, frameworkCache, urlName(energyUrl)), frameworkPath: c.Install.Path, url: energyUrl}
		for key, dl := range downloads {
			fmt.Printf("start download %s url: %s\n", key, dl.url)
			bar := progressbar.NewBar(100)
			bar.SetNotice("downloading " + dl.fileName + ": ")
			bar.SetGraph("█")
			bar.HideRatio()
			err = downloadFile(dl.url, dl.downloadPath, func(totalLength, processLength int64) {
				bar.PrintBar(int((float64(processLength) / float64(totalLength)) * 100))
			})
			bar.PrintEnd("download " + dl.fileName + " end")
			if err != nil {
				println("download", dl.fileName, "error", err)
			}
			dl.success = err == nil
		}
		println("Release files")
		for key, di := range downloads {
			if di.success {
				if key == cefKey {
					bar := progressbar.NewBar(0)
					bar.SetNotice("Unpack file " + di.fileName + ": ")
					tarName := UnBz2ToTar(di.downloadPath, func(totalLength, processLength int64) {
						bar.PrintSizeBar(processLength)
					})
					bar.PrintEnd()
					println("Unpack file", di.fileName, "end.")
					ExtractFiles(key, tarName, di, extractOSConfig)
				} else if key == energyKey {
					ExtractFiles(key, di.downloadPath, di, extractOSConfig)
				}
			}
		}
	} else {
		println("Invalid version number:", c.Install.Version)
		os.Exit(1)
	}
	return nil
}

//提取文件
func ExtractFiles(keyName, sourcePath string, di *downloadInfo, extractOSConfig map[string]interface{}) {
	println("extract", keyName, "sourcePath:", sourcePath, "targetPath:", di.frameworkPath)
	files := extractOSConfig[keyName].([]interface{})
	fmt.Println("extractOSConfig:", files)
	if keyName == cefKey {
		//tar
	} else if keyName == energyKey {
		//zip
		ExtractUnZip(sourcePath, di.frameworkPath, files...)
	}
}

func ExtractUnTar(filePath, targetPath string, files ...interface{}) {
	reader, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error: cannot read tar file, error=[%v]\n", err)
		return
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			//fmt.Printf("error: cannot read tar file, error=[%v]\n", err)
			return
		}
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(header.Name, info.Mode()); err != nil {
				//fmt.Printf("error: cannot mkdir file, error=[%v]\n", err)
				return
			}
		} else {
			if err = os.MkdirAll(path.Dir(header.Name), info.Mode()); err != nil {
				//fmt.Printf("error: cannot file mkdir file, error=[%v]\n", err)
				return
			}
			file, err := os.OpenFile(header.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				//fmt.Printf("error: cannot open file, error=[%v]\n", err)
				return
			}
			defer file.Close()
			fmt.Println("Extract file: ", header.Name)
			bar := progressbar.NewBar(100)
			bar.SetNotice("\t")
			bar.HideRatio()
			writeFile(tarReader, file, header.Size, func(totalLength, processLength int64) {
				ratio := int((float64(processLength) / float64(totalLength)) * 100)
				bar.PrintBar(ratio)
			})
			bar.PrintBar(100)
			bar.PrintEnd()
			if err != nil {
				fmt.Printf("error: cannot write file, error=[%v]\n", err)
				return
			}
		}
	}
}

func ExtractUnZip(filePath, targetPath string, files ...interface{}) {
	if rc, err := zip.OpenReader(filePath); err == nil {
		defer rc.Close()
		for i := 0; i < len(files); i++ {
			if f, err := rc.Open(files[i].(string)); err == nil {
				defer f.Close()
				st, _ := f.Stat()
				targetFileName := filepath.Join(targetPath, st.Name())
				if st.IsDir() {
					os.MkdirAll(targetFileName, st.Mode())
					continue
				}
				if targetFile, err := os.Create(targetFileName); err == nil {
					fmt.Println("Extract file: ", st.Name())
					bar := progressbar.NewBar(100)
					bar.SetNotice("\t")
					bar.HideRatio()
					writeFile(f, targetFile, st.Size(), func(totalLength, processLength int64) {
						ratio := int((float64(processLength) / float64(totalLength)) * 100)
						bar.PrintBar(ratio)
					})
					bar.PrintBar(100)
					bar.PrintEnd("\n")
					targetFile.Close()
				}
			} else {
				fmt.Printf("error: cannot open file, error=[%v]\n", err)
				return
			}
		}
	} else {
		if err != nil {
			fmt.Printf("error: cannot read zip file, error=[%v]\n", err)
		}
	}
}

//释放bz2文件到tar
func UnBz2ToTar(name string, callback func(totalLength, processLength int64)) string {
	fileBz2, err := os.Open(name)
	if err != nil {
		fmt.Errorf("%s", err.Error())
		os.Exit(1)
	}
	defer fileBz2.Close()
	dirName := fileBz2.Name()
	dirName = dirName[:strings.LastIndex(dirName, ".")]
	r := bzip2.NewReader(fileBz2)
	w, err := os.Create(dirName)
	if err != nil {
		fmt.Errorf("%s", err.Error())
		os.Exit(1)
	}
	defer w.Close()
	writeFile(r, w, 0, callback)
	return dirName
}

func writeFile(r io.Reader, w *os.File, totalLength int64, callback func(totalLength, processLength int64)) {
	buf := make([]byte, 1024)
	var written int64
	for {
		nr, err := r.Read(buf)
		if nr == 0 || err != nil {
			break
		}
		nw, err := w.Write(buf[:nr])
		if nw > 0 {
			written += int64(nw)
		}
		if callback != nil {
			callback(totalLength, written)
		}
		if err != nil {
			break
		}
		if nr != nw {
			err = io.ErrShortWrite
			break
		}
	}
}

//url文件名
func urlName(downloadUrl string) string {
	if u, err := url.QueryUnescape(downloadUrl); err != nil {
		return ""
	} else {
		u = u[strings.LastIndex(u, "/")+1:]
		return u
	}
}

//下载文件配置
func downloadConfig(url string) ([]byte, error) {
	client := new(http.Client)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func isFileExist(filename string, filesize int64) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if filesize == info.Size() {
		return true
	}
	del := os.Remove(filename)
	if del != nil {
		println(del)
	}
	return false
}

//下载文件
func downloadFile(url string, localPath string, callback func(totalLength, processLength int64)) error {
	var (
		fsize int64
	)
	tmpFilePath := localPath + ".download"
	client := new(http.Client)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	fsize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		return err
	}
	if isFileExist(localPath, fsize) {
		return err
	}
	println("save path: [", localPath, "] file size:", fsize)
	file, err := os.Create(tmpFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if resp.Body == nil {
		return errors.New("body is null")
	}
	defer resp.Body.Close()
	writeFile(resp.Body, file, fsize, callback)
	if err == nil {
		file.Close()
		err = os.Rename(tmpFilePath, localPath)
		if err != nil {
			return err
		}
	}
	return err
}
