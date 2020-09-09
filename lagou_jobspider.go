package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

type LagouJobSpider struct {
	JobSpider
}

func getHttpRequest(url string) *http.Request {
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 9_0 like Mac OS X) AppleWebKit/600.1.3 (KHTML, like Gecko) Version/8.0 Mobile/12A4345d Safari/600.1.4")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-encoding", "deflate, br")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("origin", "https://www.lagou.com")
	req.Header.Set("referer", "https://www.lagou.com/jobs/list_Java/?city=深圳")
	return req
}

func (obj *LagouJobSpider) FetchJSON(httpUrl string, pageStart int) {
	client.Get("https://www.lagou.com/jobs/list_Java/?city=%E6%B7%B1%E5%9C%B3")
	req := getHttpRequest(httpUrl)
	values := url.Values{}
	values.Set("pn", strconv.Itoa(pageStart))
	values.Set("kd", "Java")
	values.Set("needAddtionalResult", "false")
	values.Set("first", "false")
	values.Set("sid", "a82c11cae22946beaaab070988753b0a")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Body = ioutil.NopCloser(strings.NewReader(values.Encode()))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("failed to connect http", err)
	}
	bufReq, _ := httputil.DumpRequest(resp.Request, true)
	fmt.Printf("DumpRequest\n%s", bufReq)
	bufRes, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("DumpResponse\n%s", bufRes)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read http", err)
	}
	hasData := obj.AddFromJson(body)
	if hasData {
		// 防止限ip
		if pageStart > 1 && pageStart%10 == 0 {
			time.Sleep(time.Second * 60)
		}
		obj.FetchJSON(httpUrl, pageStart+1)

		fmt.Printf("page\n%s", pageStart)
	}
}

func (obj *LagouJobSpider) AddFromJson(body []byte) (hasData bool) {
	list := gjson.GetBytes(body, "content.positionResult.result")
	db := getDb()
	defer db.Close()
	for _, item := range list.Array() {
		hasData = true
		//company
		var company = new(Company)
		company.OutId = int(item.Get("companyId").Int())
		company.Source = JobSourceLagou
		if db.Where(&company).First(&company).RecordNotFound() {
			company.Name = item.Get("companyFullName").String()
			company.Site = fmt.Sprintf("https://www.lagou.com/gongsi/%s.html", item.Get("companyId").String())
			fmt.Printf("company add:%#v\n", company)
			db.Create(company)
		}
		// job
		var job = new(Job)
		job.OutId = int(item.Get("positionId").Int())
		job.Source = JobSourceLagou
		pubTime := parseLagouTime(item.Get("createTime").String())
		if db.Where(&job).First(&job).RecordNotFound() {
			job.Name = item.Get("positionName").String()
			job.Site = fmt.Sprintf("https://www.lagou.com/jobs/%s.html", item.Get("positionId").String())
			job.Address = item.Get("city").String() + "." + item.Get("district").String()
			job.PubTime = pubTime
			job.KeyWord = item.Get("hitags").String()
			job.Descr = item.Get("industryField").String() + "  " + item.Get("financeStage").String()
			salary := digitsRegexp.FindAllString(item.Get("salary").String(), -1)
			job.CompanyId = company.Id
			if len(salary) >= 1 {
				job.SalaryMin, _ = strconv.Atoi(salary[0])
				job.SalaryMin = job.SalaryMin * 1000
			}
			if len(salary) >= 2 {
				job.SalaryMax, _ = strconv.Atoi(salary[1])
				job.SalaryMax = job.SalaryMax * 1000
			}
			if len(salary) >= 3 {
				log.Println(salary)
			}
			db.Create(job)
			fmt.Printf("job add:%#v\n", job)
		} else {
			db.Model(&job).Update("PubTime", pubTime)
		}
	}
	fmt.Printf("%s", list)
	return
}

func (obj *LagouJobSpider) Fetch(httpUrl string, pageStart int) {
	u := strings.Replace(httpUrl, "%d", strconv.Itoa(pageStart), 1)
	log.Println("fetch:" + u)
	resp, err := client.Do(getHttpRequest(u))
	if err != nil {
		log.Fatal("failed to connect http", err)
	}

	bufReq, _ := httputil.DumpRequest(resp.Request, true)
	fmt.Printf("DumpRequest\n%s", bufReq)
	bufRes, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("DumpResponse\n%s", bufRes)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read http", err)
	}
	hasData := obj.AddFromHtml(body)
	if hasData {
		time.Sleep(time.Second * 2)
		obj.Fetch(httpUrl, pageStart+1)
	}
}

func (obj *LagouJobSpider) AddFromHtml(body []byte) (hasData bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("bad ret:%s\n", body)
	}
	db := getDb()
	defer db.Close()
	resultList := doc.Find("#s_position_list .item_con_list")
	ht, err := resultList.Html()
	log.Printf("resultList:%v", ht)
	resultList.Find(".con_list_item").Each(func(i int, ul *goquery.Selection) {
		site, exists := ul.Find(".company_name a").Attr("href")
		if !exists {
			return
		}
		hasData = true
		//company
		var company = new(Company)
		outId, _ := ul.Find(".company_name a").Attr("data-lg-tj-cid")
		company.OutId, _ = strconv.Atoi(outId)
		company.Source = JobSourceLagou
		if db.Where(&company).First(&company).RecordNotFound() {
			company.Name = strings.TrimSpace(ul.Find(".company_name a").Text())
			company.Site = site
			fmt.Printf("company add:%#v\n", company)
			db.Create(company)
		}
		// job
		jobId, exists := ul.Find(".p_top a").Attr("data-lg-tj-cid")
		var job = new(Job)
		job.OutId, _ = strconv.Atoi(jobId)
		job.Source = JobSourceLagou
		pubTime := parseLagouTime(ul.Find(".p_top .format-time").Text())
		if db.Where(&job).First(&job).RecordNotFound() {
			job.Name = strings.TrimSpace(ul.Find(".p_top h3").Text())
			job.Site, _ = ul.Find(".p_top a").Attr("href")
			job.Address = ul.Find(".p_top span em").Text()
			job.PubTime = pubTime
			job.KeyWord = ul.Find(".list_item_bot .li_b_l").Text()
			job.Descr = strings.TrimSpace(ul.Find(".industry").Text()) + "  " + strings.TrimSpace(ul.Find(".list_item_bot .li_b_r").Text())
			salary := digitsRegexp.FindAllString(ul.Find(".money").Text(), -1)
			job.CompanyId = company.Id
			if len(salary) >= 1 {
				job.SalaryMin, _ = strconv.Atoi(salary[0])
				job.SalaryMin = job.SalaryMin * 1000
			}
			if len(salary) >= 2 {
				job.SalaryMax, _ = strconv.Atoi(salary[1])
				job.SalaryMax = job.SalaryMax * 1000
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

func parseLagouTime(t string) time.Time {
	dateTimeFormat := "2006-01-02 15:04:05"
	dayFormat := "2006-01-02"
	ret := time.Now()
	if r1 := regexp.MustCompile(`\d\d\:\d\d发布`).FindAllString(t, -1); len(r1) > 0 {
		ret, _ = time.Parse(dateTimeFormat, ret.Format(dayFormat)+" "+r1[0][0:5]+":00")
	} else if r1 := regexp.MustCompile(`\d+天前发布`).FindAllString(t, -1); len(r1) > 0 {
		day := regexp.MustCompile(`\d+`).FindAllString(t, -1)[0]
		numDay, _ := strconv.Atoi(day)
		ret = ret.AddDate(0, 0, -numDay).Truncate(time.Hour * 24).Add(time.Hour * -8)
	} else if len(t) == len(dateTimeFormat) {
		ret, _ = time.Parse(dateTimeFormat, t)
	} else {
		ret = ret.Truncate(time.Hour)
	}
	return ret
}
