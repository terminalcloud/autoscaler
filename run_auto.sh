
cd $( dirname "${BASH_SOURCE[0]}" )

rm autoscaler
go build .

##PUT USER AND AUTH TOKEN HERE IN PLACE OF THESE SAMPLE CREDS##
USER_TOKEN=something
ACCESS_TOKEN=something
URL=http://www.somedomain.com/api/v0.2/
# don't shrink for 1 hour after starting
TIME_TO_SHRINK=3600
NODE_TYPE=g2.2xlarge
NODE_TYPE_RAM=15000

./autoscaler -usertoken=$UT -accesstoken=$AT -apiurl=$URL -policy=gpu -tts=$TIME_TO_SHRINK -nodetype=$NODE_TYPE -nodetyperam=$NODE_TYPE_RAM
