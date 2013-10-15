/*
		Copyright (c) 2008 - 2013, yourchanges.tk/blog
		All rights reserved.
		作者: 李远军 (yourchanges@gmail.com)

	gensite是一个命令行工具，主要用于读取给定的txt文本文件（通常是小说），经过解析后，生成一个小说静态站点。

	该工具会读取当前目录下的conf.ini配置文件，同时基于article_template.html和index_template.html两个模板进行相应的静态页面的生成。


*/
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/astaxie/beego/config"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	AppPath               string
	AppConfigPath         string
	AppConfig             config.ConfigContainer
	SiteURL               string
	SiteName              string
	SiteKeywords          string
	SiteDescription       string
	SiteUjianVerification string
	IndexTemplatePath     string
	ArticleTemplatePath   string
	SourceFilePath        string
	FilterDreck           string
)

type Site struct {
	SiteURL               string
	SiteName              string
	SiteKeywords          string
	SiteDescription       string
	SiteUjianVerification string
}

type PageNav struct {
	PagePrev int
	PageNext int
}

type Article struct { // 文章
	Title   string    // 文章标题
	Head    string    // 文章Head
	Content []string  // 文章内容
	Pubdate time.Time // 发布日期
	Site
	PageNav
}

//初始化
func init() {
	os.Chdir(path.Dir(os.Args[0]))
	AppPath = path.Dir(os.Args[0])
	AppConfigPath = path.Join(AppPath, "conf.ini")
	IndexTemplatePath = path.Join(AppPath, "template", "index_template.html")
	ArticleTemplatePath = path.Join(AppPath, "template", "article_template.html")
	ParseConfig()
}

//解析配置文件
func ParseConfig() (err error) {
	AppConfig, err = config.NewConfig("ini", AppConfigPath)
	if err != nil {
		return err
	} else {
		SiteURL = AppConfig.String("SiteURL")
		SiteName = AppConfig.String("SiteName")
		SiteKeywords = AppConfig.String("SiteKeywords")
		SiteDescription = AppConfig.String("SiteDescription")
		SiteUjianVerification = AppConfig.String("SiteUjianVerification")
		SourceFilePath = AppConfig.String("SourceFilePath")
		FilterDreck = AppConfig.String("FilterDreck")
		//if v, err := AppConfig.Int("httpport"); err == nil {
		//	HttpPort = v
		//}
	}
	return nil
}

//判断异常，如果有异常结束程序
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//写具体某篇文章页面
func writeArticle(article Article, index int) {
	article.PagePrev = index - 1
	article.PageNext = index + 1

	funcs := template.FuncMap{"gt": Gt}
	//fmt.Println(funcs)
	templateName := path.Base(ArticleTemplatePath)
	t, err := template.New(templateName).Funcs(funcs).ParseFiles(ArticleTemplatePath)
	checkErr(err)
	sbuffer := bytes.NewBufferString("")
	if t != nil {
		err = t.ExecuteTemplate(sbuffer, templateName, article)
		checkErr(err)
		//fmt.Println("here")
	}

	//fmt.Println(sbuffer.String())

	file, err := os.OpenFile(path.Join(AppPath, "out", strconv.Itoa(index)+".html"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	checkErr(err)
	defer file.Close()

	file.WriteString(sbuffer.String())

	//c <- 1
}

//写index.html
func writeIndex(indexPage Article, index int) {

	funcs := template.FuncMap{"add": Add}
	//fmt.Println(funcs)
	templateName := path.Base(IndexTemplatePath)
	t, err := template.New(templateName).Funcs(funcs).ParseFiles(IndexTemplatePath)
	//fmt.Println(t)
	checkErr(err)
	sbuffer := bytes.NewBufferString("")
	if t != nil {
		//t.Execute(sbuffer, indexPage)
		err = t.ExecuteTemplate(sbuffer, templateName, indexPage)
		checkErr(err)
		//fmt.Println("here")
	}

	//fmt.Println(sbuffer.String())

	file, err := os.OpenFile(path.Join(AppPath, "out", "index.html"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	checkErr(err)
	defer file.Close()

	file.WriteString(sbuffer.String())

	//c <- 1
}

// writeLines writes the lines to the given file.
//func writeLines(lines []string, path string) error {
//	file, err := os.Create(path)
//	if err != nil {
//		return err
//	}
//	defer file.Close()

//	w := bufio.NewWriter(file)
//	for _, line := range lines {
//		fmt.Fprintln(w, line)
//	}
//	return w.Flush()
//}

//模版函数 add
func Add(x, y int) int {
	return x + y
}

//模版函数 gt
func Gt(x, y int) bool {
	return x > y
}

//过滤广告
func filterDreck(line string) string {
	r := line
	if strings.Contains(line, FilterDreck) {
		r = strings.Replace(line, FilterDreck, SiteURL+" "+SiteName, -1)
	}
	return r
}

//处理函数
func readLines(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	//var lines []string
	scanner := bufio.NewScanner(file)
	site := Site{SiteURL, SiteName, SiteKeywords, SiteDescription, SiteUjianVerification}
	pageNav := PageNav{0, 0}
	indexPage := Article{SiteName + "---首页", "首页", make([]string, 0), time.Now(), site, pageNav}
	//var article []string
	article := Article{"", "", make([]string, 0), time.Now(), site, pageNav}

	intoArticle := false
	i := 1
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		line = filterDreck(line)
		if strings.HasPrefix(line, "第") && strings.Contains(line, "卷 ") && strings.Contains(line, "章 ") {
			if len(article.Content) == 0 {
				intoArticle = true
				article.Title = SiteName + "---" + line
				article.Head = line

			} else {
				//处理上一章
				writeArticle(article, i)
				//fmt.Println(article)
				i = i + 1

				//新建容器，来放新的一章
				article = Article{SiteName + "---" + line, line, make([]string, 0), time.Now(), site, pageNav}
				article.Content = append(article.Content, line)
				//break
			}

			//用于生成index.html
			indexPage.Content = append(indexPage.Content, line)

		} else {
			if intoArticle {
				article.Content = append(article.Content, line)
			}
		}

	}
	if len(article.Content) > 0 {
		//处理上一章
		writeArticle(article, i)
	}

	if len(indexPage.Content) > 0 {
		//生成index.html
		writeIndex(indexPage, i)
	}
	return scanner.Err()
}

//复制并覆盖文件
func CopyFile(src, dst string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

//主函数
func main() {
	fmt.Println("Starting")
	err := readLines(SourceFilePath)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	_, err = CopyFile(path.Join(AppPath, "template", "base.css"), path.Join(AppPath, "out", "base.css"))
	if err != nil {
		log.Fatalf("copy file: %s", err)
	}
	fmt.Println("Finished")
	//for i, line := range lines {
	//	fmt.Println(i, line)
	//}
}
