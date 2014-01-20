package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type User struct {
	Username string
	Password string
	Wwwroot  string
}
type Users struct {
	Users []User
}

var (
	users   Users
	wwwroot string
)

const (
	VERSION     = "1.0"
	SERVER_INFO = "TjzAimee Picture System v" + VERSION
)

func parseConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
		return
	}
	buff := make([]byte, 2048)
	n, err := file.Read(buff)
	if err != nil {
		panic(err)
		return
	}
	err = json.Unmarshal(buff[:n], &users)
	if err != nil {
		panic(err)
		return
	}
}
func init() {
	parseConfig()
}

func dirExists(name string) (bool, error) {
	fileInfo, err := os.Stat(name)
	if err == nil {
		return fileInfo.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func fileExists(name string) (bool, error) {
	fileInfo, err := os.Stat(name)
	if err == nil {
		return !fileInfo.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func validate(rw http.ResponseWriter, req *http.Request) bool {
	rw.Header().Set("Server", SERVER_INFO)
	username := req.FormValue("username")
	log.Println("username = " + username)
	password := req.FormValue("password")
	/*log.Println("password = " + password)*/
	for _, v := range users.Users {
		if username == v.Username && password == v.Password {
			wwwroot = "./" + v.Wwwroot + "/"
			return true
		}
	}
	return false
}
func upload(rw http.ResponseWriter, req *http.Request) {
	if !validate(rw, req) {
		http.NotFound(rw, req)
		return
	}
	if "POST" == req.Method {
		filename := req.FormValue("filename")
		filepath := req.FormValue("filepath")
		file, _, err := req.FormFile("file")
		if err != nil {
			fmt.Fprintln(rw, `{"errcode":40004,"errmsg":"Uploaded file not found."}`)
			panic(err)
			return
		}
		buff, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Fprintln(rw, `{"errcode":40003,"errmsg":"File read failed."}`)
			panic(err)
			return
		}
		dirName := wwwroot + "/" + filepath + "/"
		dirFlag, _ := dirExists(dirName)
		if !dirFlag {
			err = os.MkdirAll(dirName, 0666)
			if err != nil {
				fmt.Fprintln(rw, `{"errcode":40002,"errmsg":"Failed to create directory."}`)
				panic(err)
				return
			}
		}
		dstFile, err := os.OpenFile(dirName+filename, os.O_WRONLY|os.O_CREATE, 0666)
		log.Println(dstFile.Name())
		if err != nil {
			fmt.Fprintln(rw, `{"errcode":40001,"errmsg":"Failed to create file."}`)
			panic(err)
			return
		}
		defer dstFile.Close()
		dstFile.Write(buff)
		fmt.Fprintln(rw, `{"errcode":0,"errmsg":"ok"}`)
	}
}
func static(rw http.ResponseWriter, req *http.Request) {
	log.Println(req.URL.Path)
	if !validate(rw, req) {
		http.NotFound(rw, req)
		return
	}
	if strings.HasPrefix(req.URL.Path, "/") {
		file := wwwroot + req.URL.Path[len("/"):]
		fileInfo, err := os.Stat(file)
		if err != nil {
			panic(err)
			return
		}
		if !fileInfo.IsDir() {
			rw.Header().Set("contentType", "application/octet-stream")
			rw.Header().Set("Content-Disposition", "attachment; filename=\""+fileInfo.Name()+"\"")
		}
		http.ServeFile(rw, req, file)
	}
}
func api(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Server", SERVER_INFO)
	rw.Header().Set("Content-Type","text/html;charset=UTF-8")
	fmt.Fprintln(rw,`
		<style type="text/css">
			*{
				padding:0;
				margin:0 auto;
				font-size:12px;
			}
			body{
				margin-top: 10px;
				width: 960px;
			}
			table thead tr{
				line-height: 27px;
			}
			table tbody tr{
				line-height: 21px;
			}
			table td{
				padding-left: 5px;
			}
			pre {
				background-color: #F8F8F8;
				border: 1px solid #CCCCCC;
				border-radius: 3px;
				color: #333333;
				font: 12px/20px "MicroSoft YaHei","Courier New","Andale Mono",monospace;
				margin-left: 10px;
				padding: 5px 10px;
				white-space: pre-wrap;
				word-wrap: break-word;
			}
		</style>
		<h5>全局返回码说明如下：</h5><br/>
  <table style="width:100%;border-collapse:collapse;" border="1">
	<thead>
		<tr>
			<th>返回码</th>
			<th>说明</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>0</td>
			<td>上传成功</td>
		</tr>
		<tr>
			<td>40001</td>
			<td>创建文件失败</td>
		</tr>
		<tr>
			<td>40002</td>
			<td>创建目录失败</td>
		</tr>
		<tr>
			<td>40003</td>
			<td>读取上传文件失败</td>
		</tr>
		<tr>
			<td>40004</td>
			<td>未找到上传的文件</td>
		</tr>
	</tbody>
  </table>
  <div>
	<br/><h5>正确时的返回JSON数据包如下：</h5>
	<pre>{"errcode":0,"errmsg":"ok"}</pre>
  </div>
  <div>
	<br/><h5>错误时的返回JSON数据包如下：</h5>
	<pre>{"errcode":40001,"errmsg":"Failed to create file."}</pre>
  </div>
   <div>
	<br/><h5>上传API：</h5>
	<pre>http://127.0.0.1:2070/upload/
		参数：
		username 用户名
		password 密码
		filepath 上传至哪个目录
		filename 文件名
		file 上传文件的File对象
	</pre>
  </div>
  <div>
	<br/><h5>下载API：</h5>
	<pre>http://127.0.0.1:2070/目录名/文件名
		参数：
		username 用户名
		password 密码
	</pre>
  </div>
	`)
}
func main() {
	fmt.Println("Picture System\nListening port:2070")
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/", static)
	http.HandleFunc("/upload/", upload)
	http.HandleFunc("/api/", api)
	err := http.ListenAndServe(":2070", nil)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
