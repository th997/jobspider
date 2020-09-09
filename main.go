package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	//_ "github.com/go-sql-driver/mysql"
	"github.com/martini-contrib/render"
)

var cjol CjolJobSpider
var qcwyJobSpider QcwyJobSpider
var lagouJobSpider LagouJobSpider
var bossJobSpider BossJobSpider

func getDb() (db *gorm.DB) {
	db, err := gorm.Open("sqlite3", "./job.db")
	//db, err := gorm.Open("mysql", "test:123456@(localhost:3306)/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("failed to connect database", err)
	}
	db.LogMode(true)
	return
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
	//初始化数据库
	db := getDb()
	defer db.Close()
	// Migrate the schema
	db.AutoMigrate(Job{}, Company{})
	db.LogMode(true)
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",       // Specify what path to load the templates from.
		Extensions: []string{".html"}, // Specify extensions to load for templates.
		Charset:    "UTF-8",           // Sets encoding for json and html content-types. Default is "UTF-8".
		IndentJSON: true,              // Output human readable JSON
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", "jeremy")
	})
	m.Get("/downloadCompany", downloadCompany)
	m.Get("/downloadCjol", downloadCjol)
	m.Get("/downloadQcwy", downloadQcwy)
	m.Get("/downloadLagou", downloadLagou)
	m.Get("/downloadBoss", downloadBoss)

	go func() {
		//cmd := exec.Command("cmd", "/c start http://127.0.0.1:13520")
		cmd := exec.Command("xdg-open", "http://127.0.0.1:13520")
		cmd.Start()
	}()
	m.RunOnAddr(":13520")
}
func downloadCjol() string {
	go func() {
		cjolUrl := "http://s.cjol.com/service/joblistjson.aspx?KeywordType=3&RecentSelected=41&KeyWord=java&Location=2008&SearchType=3&ListType=2&Sortings=MinSalary%20desc&page=%d"
		cjol.Fetch(cjolUrl, 1)
	}()
	return "ok"
}
func downloadQcwy() string {
	go func() {
		qianchengwuyouUrl := `http://search.51job.com/list/040000,000000,0000,00,9,99,java,2,%d.html?lang=c&stype=1&postchannel=0000&workyear=99&cotype=99&degreefrom=99&jobterm=99&companysize=99&lonlat=0%2C0&radius=-1&ord_field=0&confirmdate=9&fromType=&dibiaoid=0&address=&line=&specialarea=00&from=&welfare=`
		qcwyJobSpider.Fetch(qianchengwuyouUrl, 1)
	}()
	return "ok"
}
func downloadLagou() string {
	go func() {
		lagouUrl := `https://www.lagou.com/jobs/positionAjax.json?city=深圳&needAddtionalResult=false`
		lagouJobSpider.FetchJSON(lagouUrl, 1)
	}()
	return "ok"
}
func downloadBoss() string {
	go func() {
		bossUrl := `https://www.zhipin.com/c101280600-p100101/?page=%d`
		bossJobSpider.Fetch(bossUrl, 1)
	}()
	return "ok"
}
func downloadCompany() string {
	go func() {
		db := getDb()
		defer db.Close()
		//db.Where("descr is null").Find(())
	}()
	return "ok"
}
