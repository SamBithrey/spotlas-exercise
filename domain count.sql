SELECT * 
FROM (SELECT distinct upper(name) as name, domain, COUNT(domain) over(partition by (upper(name), domain)) as "count number"
	  FROM (SELECT name, substring(website from '(?:.*://)?(?:www\.)?([^/?]*)') as domain
			FROM "MY_TABLE") a) a
WHERE "count number" > 1
ORDER BY 2;
