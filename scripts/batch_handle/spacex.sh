#!/bin/bash

# shell 模板
>/tmp/spacex.log
bin=`dirname "$0"`
bin=`cd "$bin"; pwd`
cmd="mysql -uadbase -pAdbase83jS2kxFl#xPQ4b -hbdb.master.adbase.niceprivate.com -P 3307 "
table="ad_base.ad_user_info"
step=1000

version=`date +"%s"`
#删除四天前的过期user info数据
delver=`expr $version - 345600`
echo $version
echo $delver

source /etc/bashrc

hdfs dfs -rm -r /tmp/spacex_user_profile
[ $? -ne 0 ] && echo "hdfs rm spacex_user_profile failure " && exit 127

hive -f $bin/spacex.hql
[ $? -ne 0 ] && echo "execute hive hql" && exit 127

rm -rf $bin/spacex.out

hdfs dfs -get /tmp/spacex_user_profile $bin/spacex.out
[ $? -ne 0 ] && echo "get spacex result from hdfs failure" && exit 127

cat $bin/spacex.out/* | python $bin/spacex.py $bin/nice.conf $version
[ $? -ne 0 ] && echo "update spacex user info to mysql db failure" && exit 127

#删除ad_user_info_0~99表的历史数据
for i in `seq 0 99`
do
	count=`$cmd -N -e "select count(*) from ${table}_${i} where version < ${delver}"`
	loop=`expr $count / $step`
	#根据步长来打散删除数据，每删除一次做一次 sleep 0.5
	for l in `seq 0 $loop`
	do
		sql="delete from ${table}_${i} where version < ${delver} limit $step"
		echo $sql
		$cmd -vve "$sql"
		sleep 0.5
	done
done


