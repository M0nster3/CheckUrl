package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)
var(
	file string
	out string
	results = make(chan int)
	Lines [][]interface{}
)
func Usage(){
	banner := `

 ______  __  __  ______  ______  __  __   __  __  ______  __        
/\  ___\/\ \_\ \/\  ___\/\  ___\/\ \/ /  /\ \/\ \/\  == \/\ \       
\ \ \___\ \  __ \ \  __\\ \ \___\ \  _"-.\ \ \_\ \ \  __<\ \ \____  
 \ \_____\ \_\ \_\ \_____\ \_____\ \_\ \_\\ \_____\ \_\ \_\ \_____\ 
  \/_____/\/_/\/_/\/_____/\/_____/\/_/\/_/ \/_____/\/_/ /_/\/_____/


Usage: CheckUrl -i file

Options:
`
	print(banner)
	flag.PrintDefaults()

}
func init() {

	flag.StringVar(&file,"i", "", "请输入检测文件")
	flag.StringVar(&out,"o", "", "请输入输出文件")
	flag.Usage = Usage
}

func outfile (filee string,datas string){
	data := []byte(datas)
	// 检测文件是否存在
	if _, err := os.Stat(filee); err == nil {
		// 文件存在，打开文件并追加数据
		f, err := os.OpenFile(filee, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		if _, err := f.Write(data); err != nil {
			return
		}
	} else {
		// 文件不存在，创建文件并写入数据
		fileee,_:=os.Create(filee)
		defer fileee.Close()
		buff:=bufio.NewWriter(fileee)
		buff.WriteString(datas+"\n")
	}

}

func HttpParse(line string){

	var wg sync.WaitGroup
	if line != "" && strings.Contains(line,"."){
		//禁止检测证书
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Timeout: 3 * time.Second, Transport: tr}
		wg.Add(1)
		resp, err := client.Get(line)

		if err !=nil && strings.Contains(err.Error(), "Client.Timeout") {
			one := []interface{}{line,"Time Out"}
			Lines=append(Lines, one)
			<-results
			return
		}
		if err !=nil  {
			one := []interface{}{line}
			Lines=append(Lines, one)
			<-results
			return
		}

		defer resp.Body.Close()
		dataBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}else {
			re := regexp.MustCompile("<title>(.*?)</title>")
			body := string(dataBytes)
			title := re.FindAllStringSubmatch(body, -1)

			if len(title)!=0{
				three := []interface{}{line,resp.StatusCode,len(dataBytes),title[0][1]}
				Lines=append(Lines, three)
				wg.Done()
				<-results
			}else {
				two := []interface{}{line,resp.StatusCode,len(dataBytes)}
				Lines=append(Lines, two)
				wg.Done()
				<-results
			}
		}
	}else {
		<-results
		return
	}
	wg.Wait()
}


func main() {
	flag.Parse()
	if file != ""{
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		fmt.Println(color.LightBlue.Sprintf("\nNumber of URLS:%d\n", len(lines)))
		for i:=0;i<=len(lines)-1;i++{
			line := lines[i]
			if len(line) >= 7 && line[:7] != "http://" && line[:8] != "https://"{
				line = "http://"+line
				go HttpParse(line)
			}else {
				go HttpParse(line)
			}
		}
		for i:=0;i<=len(lines)-1;i++{
			results <-i
		}

		for _,a :=range Lines{
			if len(a)>=2{
				code := fmt.Sprintf("%v", a[1])
				if strings.Contains(code, "200") {
					if len(a) == 3{
						fmt.Println(color.LightCyan.Sprintf("%-80s", a[0]) + color.LightGreen.Sprintf(" [ code: %d, Bytes: %d ]\n", a[1], a[2]))
						if out !=""{
							code1 := fmt.Sprintf("%v", a[0])
							code2 := fmt.Sprintf("%v", a[1])
							code3 := fmt.Sprintf("%v", a[2])
							outfile(out,"[URL:"+code1+"code:"+code2+"Bytes:"+code3+"]\n")
						}
					}else if len(a) == 4{
						fmt.Println(color.LightCyan.Sprintf("%-80s", a[0]) + color.LightGreen.Sprintf(" [ code: %d, Bytes: %d, title: %s ]\n", a[1], a[2], a[3]))
						if out !=""{
							code1 := fmt.Sprintf("%v", a[0])
							code2 := fmt.Sprintf("%v", a[1])
							code3 := fmt.Sprintf("%v", a[2])
							code4 := fmt.Sprintf("%v", a[3])
							outfile(out,"[URL:"+code1+"code:"+code2+"Bytes:"+code3+"Title:"+code4+"]\n")
						}
					}
				}
			}
		}

		for _,a :=range Lines {
			if len(a) > 2 {
				code := fmt.Sprintf("%v", a[1])

				if !strings.Contains(code, "200") && len(a) != 1 {
					if len(a) == 3 {
						fmt.Println(color.LightCyan.Sprintf("%-80s", a[0]) + color.LightRed.Sprintf("[ code: %d, Bytes: %d ]\n", a[1], a[2]))
						if out !=""{
							code1 := fmt.Sprintf("%v", a[0])
							code2 := fmt.Sprintf("%v", a[1])
							code3 := fmt.Sprintf("%v", a[2])
							outfile(out,"[URL:"+code1+"code:"+code2+"Bytes:"+code3+"]\n")
						}
					}
					if len(a) == 4 {
						fmt.Println(color.LightCyan.Sprintf("%-80s", a[0]) + color.LightRed.Sprintf("[ code: %d, Bytes: %d. Title: %s ]\n", a[1], a[2],a[3]))
						if out !=""{
							code1 := fmt.Sprintf("%v", a[0])
							code2 := fmt.Sprintf("%v", a[1])
							code3 := fmt.Sprintf("%v", a[2])
							code4 := fmt.Sprintf("%v", a[3])
							outfile(out,"[URL:"+code1+"code:"+code2+"Bytes:"+code3+"Title:"+code4+"]\n")
						}
					}
				}
			}
		}
		for _,a :=range Lines {
			if len(a) == 2 {
				code := fmt.Sprintf("%v", a[1])
				if strings.Contains(code, "Time") {
					fmt.Println(color.LightCyan.Sprintf("%-80s", a[0]) + color.LightMagenta.Sprintf("[ %s ]\n",a[1]))
					if out !=""{
						code1 := fmt.Sprintf("%v", a[0])
						code2 := fmt.Sprintf("%v", a[1])
						outfile(out,"[URL:"+code1+""+code2+"]")
					}
				}
			}
		}
		for _,a :=range Lines{
			if len(a)==1{
				fmt.Println(color.LightCyan.Sprintf("%-80s", a[0]) +  color.LightYellow.Sprintf(" [ Error ]\n"))
				if out !=""{
					code1 := fmt.Sprintf("%v", a[0])
					outfile(out,"[URL:"+code1+" Error]")
				}
			}
		}
	}else {
		Usage()
	}
}