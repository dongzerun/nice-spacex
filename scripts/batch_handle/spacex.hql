set hive.exec.compress.output=false;
insert overwrite directory '/tmp/spacex_user_profile'
select t5.uid, t5.name, gender, age, channel, ctime, pf, tags from 
(select id AS uid, name, gender, age, utm_source AS channel, ctime from kkgoo.kk_user) t5
left outer join 
(select uid, collect_set(platform) AS pf from kkgoo.kk_user_bind_account group by uid) t4
on t5.uid=t4.uid 
left outer join
(select t2.uid, collect_set(t3.name) AS tags from 
(select uid, tagid, type from kkgoo.kk_user_show_image_tag_idx where type in ('exists', 'undefined') group by uid, tagid, type) t2
left outer join 
(select tagid, type, name from (
select id AS tagid, 'exists' AS type, name from kkgoo.kk_brand_information
union all 
select id AS tagid, 'undefined' AS type, name from kkgoo.kk_user_show_brand_tag) t1) t3
on t2.tagid=t3.tagid and t2.type=t3.type 
group by t2.uid) t6
on t6.uid=t5.uid
