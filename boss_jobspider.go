package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
)

type BossJobSpider struct {
	JobSpider
}

var baseUrl = "https://www.zhipin.com/"

func bossGetHttpRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36")
	req.Header.Set("Host", "Host")
	req.Header.Set("Accept", "*/*")
	//req.Header.Set("Connection", "close")
	//req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	//req.Header.Set("accept-encoding", "deflate, br")
	//req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	//req.Header.Set("cache-control", "max-age=0")
	//req.Header.Set("dnt", "1")
	//req.Header.Set("sec-fetch-dest", "document")
	//req.Header.Set("sec-fetch-mode", "navigate")
	//req.Header.Set("sec-fetch-site", "none")
	//req.Header.Set("upgrade-insecure-requests", "1")
	//req.Header.Set("cookie","__zp_stoken__=2634aWGNqBXY2AhleIBsuSHskZEtGNXpMeXI%2BKx89JEI2K3EATWwVSGBMHzUlEEl9Fx94QH1wES1sCGVkMHZdBmtpVyt9GkltWRddD3BJRh9leGtnIXx4IhY9K1Iya3gWVSY9PwwOEFVyTWE%3D")
	return req
}

func (obj *BossJobSpider) Fetch(httpUrl string, pageStart int) {
	u := strings.Replace(httpUrl, "%d", strconv.Itoa(pageStart), 1)
	log.Println("fetch:" + u)
	//if pageStart== 1 {
	//	client.Get("https://www.zhipin.com/shenzhen/")
	//}
	req := bossGetHttpRequest(u)

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return errors.New("stopped after 1 redirects")
		}
		return nil
	}

	bufReq, _ := httputil.DumpRequest(req, true)
	fmt.Printf("DumpRequest\n%s", bufReq)

	resp, err := client.Do(req)
	if err != nil {
		locationUrl := resp.Header.Get("location")
		if locationUrl != "" {
			resp, err = client.Get(baseUrl + locationUrl)
			bufRes, _ := httputil.DumpResponse(resp, true)
			fmt.Printf("DumpResponse\n%s", bufRes)
			resp, err = client.Do(req)
		} else {
			log.Fatal("failed to connect http", err)
		}
	}
	bufRes, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("DumpResponse\n%s", bufRes)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(mahonia.NewDecoder("utf-8").NewReader(resp.Body))
	if err != nil {
		log.Fatal("failed to read http", err)
	}
	hasData := obj.AddFromHtml(body)
	if hasData {
		//time.Sleep(time.Second * 2)
		obj.Fetch(httpUrl, pageStart+1)
	}
}

/*

 */
func (obj *BossJobSpider) AddFromHtml(body []byte) (hasData bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("bad ret:%s\n", body)
	}
	db := getDb()
	defer db.Close()
	resultList := doc.Find("ul li .job-primary")
	ht, err := resultList.Html()
	log.Printf("resultList:%v", ht)
	resultList.Each(func(i int, ul *goquery.Selection) {
		site, exists := ul.Find(".company-text h3 a").Attr("href")
		if !exists {
			return
		}
		hasData = true
		//company
		var company = new(Company)
		comId, _ := ul.Find(".primary-box").Attr("data-itemid")
		company.OutId, _ = strconv.Atoi(comId)
		company.Source = JobSourceBoss
		if db.Where(&company).First(&company).RecordNotFound() {
			company.Name = strings.TrimSpace(ul.Find(".company-text h3").Text())
			company.Site = baseUrl + site
			fmt.Printf("company add:%#v\n", company)
			db.Create(company)
		}
		// job
		jobId, _ := ul.Find("div .primary-box").Attr("data-jobid")
		var job = new(Job)
		job.OutId, _ = strconv.Atoi(jobId)
		job.Source = JobSourceBoss
		pubTime := time.Now()
		if db.Where(&job).First(&job).RecordNotFound() {
			job.Name = strings.TrimSpace(ul.Find(".info-publis h3 em").Text())
			job.Site, _ = ul.Find("div .primary-box").Attr("href")
			job.Address = ul.Find(".job-area").Text()
			job.PubTime = pubTime
			salary := digitsRegexp.FindAllString(ul.Find(".job-limit .red").Text(), -1)
			job.CompanyId = company.Id
			if len(salary) >= 1 {
				job.SalaryMin, _ = strconv.Atoi(salary[0])
			}
			if len(salary) >= 2 {
				job.SalaryMax, _ = strconv.Atoi(salary[1])
			}
			if len(salary) >= 3 {
				log.Println(salary)
			}
			db.Create(job)
			fmt.Printf("job add:%#v\n", job)
		}
	})
	return
}
