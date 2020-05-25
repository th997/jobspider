package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type CjolJobSpider struct {
	JobSpider
}

// cjol http json
type CjolHtml struct {
	IsRecommend bool
	RecordSum   int
	VipPostSum  int
	JobListHtml string
}

func (obj *CjolJobSpider) Fetch(httpUrl string, pageStart int) {
	u := strings.Replace(httpUrl, "%d", strconv.Itoa(pageStart), 1)
	resp, err := http.PostForm(u, url.Values{})
	if err != nil {
		log.Fatal("failed to connect http", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read http", err)
	}
	hasData := obj.AddFromHtml(body)
	if hasData {
		obj.Fetch(httpUrl, pageStart+1)
	}
}

/* job html 格式
<ul class="results_list_box">
<li class="list_type_checkbox"><input type="checkbox" class="checkbox" ispast="False" companyid="72526" value="7484472"  /></li>
<li class="list_type_first"><h3><a href="http://www.cjol.com/jobs/job-7484472#source=101" target="_blank">Window 软件开发工程师</a></h3><em class="icon_sticktop" title="置顶职位"></em><em jobid="7484472" class="icons_collect_listpage icon_collect_listpage_click" title="收藏职位"></em></li>
<li class="list_type_second"><a href="http://www.cjol.com/jobs/company-72526" target="_blank">深圳市科瑞康实业有限公司</a></li>
<li class="list_type_third">广东深圳</li>
<li class="list_type_fifth">本科以上
<li class="list_type_sixth">0</li>
<li class="list_type_seventh">5000至10000元</li>
<li class="list_type_eighth">2016-10-19</li><i class="imitate_i"></i>
</ul>
*/
func (obj *CjolJobSpider) AddFromHtml(body []byte) (hasData bool) {
	var cjolHtml CjolHtml
	json.Unmarshal(body, &cjolHtml)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(cjolHtml.JobListHtml))
	if err != nil {
		log.Printf("bad json:%s\n", body)
	}
	db := getDb()
	defer db.Close()
	doc.Find(".results_list_box").Each(func(i int, ul *goquery.Selection) {
		hasData = true
		companyId, exists := ul.Find(".list_type_checkbox input").Attr("companyid")
		if !exists {
			fmt.Printf("ul:%s\n", ul)
		}
		var company = new(Company)
		company.OutId, _ = strconv.Atoi(companyId)
		company.Source = JobSourceCjol
		// save company
		if db.Where(&company).First(&company).RecordNotFound() {
			company.Name = ul.Find(".list_type_second").Text()
			company.Site, _ = ul.Find(".list_type_second").Find("a").Attr("href")
			db.Create(company)
		}
		jobId, exists := ul.Find(".list_type_checkbox input").Attr("value")
		if !exists {
			fmt.Printf("company:%#v\n", company)
		}
		var job = new(Job)
		job.OutId, _ = strconv.Atoi(jobId)
		job.Source = JobSourceCjol
		pubTime, _ := time.Parse("2006-01-02", ul.Find(".list_type_eighth").Text())
		// save job
		if db.Where(&job).First(&job).RecordNotFound() {
			job.Name = ul.Find(".list_type_first").Text()
			job.Site, _ = ul.Find(".list_type_first a").Attr("href")
			job.Address = ul.Find(".list_type_third").Text()
			job.Degree = ul.Find(".list_type_fifth").Text()
			job.Experience, _ = strconv.Atoi(ul.Find(".list_type_sixth").Text())
			job.PubTime = pubTime
			salary := digitsRegexp.FindAllString(ul.Find(".list_type_seventh").Text(), -1)
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
