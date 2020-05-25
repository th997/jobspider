package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
)

type QcwyJobSpider struct {
	JobSpider
}

func (obj *QcwyJobSpider) Fetch(httpUrl string, pageStart int) {
	u := strings.Replace(httpUrl, "%d", strconv.Itoa(pageStart), 1)
	log.Println("fetch:" + u)
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal("failed to connect http", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(mahonia.NewDecoder("gbk").NewReader(resp.Body))
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
<div class="el">
	<p class="t1 ">
		<input class="checkbox" type="checkbox" name="selectJobid72750858" value="72750858">
		<span><a target="_blank" title="中级JAVA开发工程师" href="http://jobs.51job.com/shenzhen-nsq/72750858.html?s=0" onmousedown="">中级JAVA开发工程师</a></span>
	</p>
	<span class="t2"><a target="_blank" title="深圳市奥创科技有限公司" href="http://jobs.51job.com/all/co2714597.html">深圳市奥创科技有限公司</a></span>
	<span class="t3">深圳-南山区</span>
	<span class="t4">8000-15000/月</span>
	<span class="t5">10-21</span>
</div>
*/
func (obj *QcwyJobSpider) AddFromHtml(body []byte) (hasData bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("bad ret:%s\n", body)
	}
	db := getDb()
	defer db.Close()
	resultList := doc.Find("#resultList")
	ht, err := resultList.Html()
	log.Printf("resultList:%v", ht)
	resultList.Find(".el").Each(func(i int, ul *goquery.Selection) {
		site, exists := ul.Find(".t2 a").Attr("href")
		if !exists {
			return
		}
		hasData = true
		//company
		var company = new(Company)
		ss := strings.Split(site, "/")
		dd := digitsRegexp.FindAllString(ss[len(ss)-1], -1)
		if len(dd) == 0 {
			fmt.Printf("company get faild:%#v\n", ul.Text())
			return
		}
		company.OutId, _ = strconv.Atoi(digitsRegexp.FindAllString(ss[len(ss)-1], -1)[0])
		company.Source = JobSource51job
		if db.Where(&company).First(&company).RecordNotFound() {
			company.Name = strings.TrimSpace(ul.Find(".t2 a").Text())
			company.Site = site
			fmt.Printf("company add:%#v\n", company)
			db.Create(company)
		}
		// job
		jobId, exists := ul.Find(".t1 input").Attr("value")
		var job = new(Job)
		job.OutId, _ = strconv.Atoi(jobId)
		job.Source = JobSource51job
		pubTime, _ := time.Parse("2006-01-02", strconv.Itoa(time.Now().Year())+"-"+ul.Find(".t5").Text())
		if db.Where(&job).First(&job).RecordNotFound() {
			job.Name = strings.TrimSpace(ul.Find(".t1 a").Text())
			job.Site, _ = ul.Find(".t1 a").Attr("href")
			job.Address = ul.Find(".t3").Text()
			job.PubTime = pubTime
			salary := digitsRegexp.FindAllString(ul.Find(".t4").Text(), -1)
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
		} else {
			db.Model(&job).Update("PubTime", pubTime)
		}
	})
	return
}
