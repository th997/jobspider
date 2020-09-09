select t2.name, t1.name,t1.address,t1.site,t1.salary_max,t1.salary_min,t1.descr,t1.pub_time from job t1
left join company t2 on t1.company_id=t2.id
 where salary_max>30000 order by t1.id desc