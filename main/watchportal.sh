#!/bin/sh

LOGFILE=/data/log/watchcatportal.log
INTERVAL=60
CMDPROGRAM=./portal_radius_server

# redirect stdout,stderr to $LOGFILE
[ -d $(dirname $LOGFILE) ] || (echo create log dir; rm -rf $(dirname $LOGFILE); mkdir -p $(dirname $LOGFILE) )
#rm -rf $LOGFILE
#exec 1>>$LOGFILE 2>&1

date "+%Y-%m-%d %H:%M:%S" >> $LOGFILE 
echo "watchcat is running . . ." >> $LOGFILE 

while true;
do
    PROGRAM=`ps -aux| grep ./portal_radius_server | grep -v grep | awk NR==1'{print $1 }'`
    #echo $PROGRAM
    if [ $PROGRAM ] ;
    then
        #PROGRAM=`ps -aux| grep relationship | grep -v grep | awk '{ print $11,$12,$13 }'`
        #CMDPROGRAM=$PROGRAM
        date "+%Y-%m-%d %H:%M:%S" >> $LOGFILE 
    else
        echo ""
        date "+%Y-%m-%d %H:%M:%S" >> $LOGFILE 
        echo "portal_radius_server not running, now restart it" >> $LOGFILE 
        $CMDPROGRAM &
    fi
    sleep $INTERVAL
   
done

date "+%Y-%m-%d %H:%M:%S" >> $LOGFILE 
echo "watchcat is exiting. . ." >> $LOGFILE 
echo "" >> $LOGFILE 

