import threading,time
import os
import xlrd
from xlutils.copy import copy;
import pymongo
import calendar
import datetime


xls_name = "./ops/bindnum/acip_bindnum.xls"
xls_path = xls_name
xls_new = xls_path.split(".xls")[0] + "_portal" +  ".xls"


class CityInfo(object):
    def __init__(self, name, village_row_start=1, village_row_end=1, village_row_step=3):
        self.name = name
        self.village_row_start = village_row_start
        self.village_row_end = village_row_end
        self.village_row_step = village_row_step


def work(city, sheet, newsheet, sheetname):
    get_rt_bind_num(city, sheet, newsheet, sheetname)
    print("work(city, sheet, newsheet, sheetname)")


def get_mongodb_client(): #省级网管 r_radius  dfwo18903822d
    mongo_user = 'rw_radius' #访问数据库的用户名
    mongo_password = 'Nu020Secu565!'  # 访问数据库的密码

    #client = pymongo.MongoClient('127.0.0.1',server.local_bind_port) ## 这里一定要填入ssh映射到本地的端口
    client = pymongo.MongoClient(host='192.168.254.11,192.168.254.12', port=30202, username=mongo_user, password=mongo_password, authSource='radius' )
    #uri = "mongodb://rw_radius:inqw13912301z@127.0.0.1:27017/radius?authSource=radius&authMechanism=MONGODB-CR"
    #client = pymongo.MongoClient(uri)
    db=client.radius #client.admin  radius
    db.authenticate(mongo_user, mongo_password)
    return client

def get_rt_bind_num(city, sheet, newsheet, sheetname):

    xls_routerip_col = 5

    xls_village_row = city.village_row_start
    xls_village_end = city.village_row_end
    xls_village_step = city.village_row_step

    print("xls_village_row,xls_village_end ,xls_village_step :", xls_village_row, xls_village_end, xls_village_step)

    xls_router_status = sheet.ncols
    xls_rt_bind_sum = xls_router_status

    #统计年月---主要获取当前时间来计数之前每月最后一天
    datetmpstr = []
    datetime_int = []
    year = datetime.datetime.now().year
    month = datetime.datetime.now().month


    if False : # 以下采集201907至今的数据  
        for i in range(7, 13):
            monthday = "%d%02d"%(2019, i)
            datetmpstr.append(monthday)
            #计数出每月最后一天
            d = calendar.monthrange(2019, i)
            datetime_int.append(datetime.datetime(2019, i, d[1], 23, 59, 59))
        if year > 2019:  
            for j in range(2020, year + 1):    
                for i in range(1, month+1):
                    monthday = "%d%02d" % (j, i)
                    datetmpstr.append(monthday)
                    #计数出每月最后一天
                    d = calendar.monthrange(j, i)
                    datetime_int.append(datetime.datetime(j, i, d[1], 23, 59, 59))
    else:
        monthday = "%d%02d" % (year, month)
        datetmpstr.append(monthday)
        #计数出每月最后一天
        d = calendar.monthrange(year, month)        
        datetime_int.append(datetime.datetime(year, month, d[1], 23, 59, 59))        

    
    #print(datetmpstr)
    #print(datetime_int)

    i = 0
    while(i < len(datetmpstr)):
        newsheet.write(0, xls_rt_bind_sum+i, datetmpstr[i])
        i = i + 1
    #统计年月---主要获取当前时间来计数之前每月最后一天

    # 再连接自己的数据库mydb
    client = get_mongodb_client()
    my_db = client.radius               # 再连接自己的数据库mydb
    collection = my_db.chs_clientinfo   # myset集合，同上解释
    area_bindcollection = my_db.area_binddata    # myset集合，同上解释
    users = client.users

    i = 0
    while (i < len(datetime_int)):
        BIND_TIME = datetime_int[i]
        xls_village_row = city.village_row_start
        xls_village_end = city.village_row_end
        binnum_total = 0
        while (xls_village_row < xls_village_end):
            acip = sheet.cell_value(xls_village_row, xls_routerip_col)

            if (not acip):
                xls_village_row = xls_village_row + xls_village_step
                continue

            mydoc = collection.aggregate([  # 查看某area某段时间eau累计绑定数
                {"$match":
                    {
                        "acIp": acip,#"113.64.224.61",
                        "bindtime": { "$lt": BIND_TIME} 
                    },
                },
                {"$count": "count"}
            ])

            #print(mydoc)
            acipbindnum = 0          
            for x in mydoc:
                acipbindnum = x["count"]

            binnum_total = binnum_total + acipbindnum
            newsheet.write(xls_village_row, xls_rt_bind_sum+i, acipbindnum)

            xls_village_row = xls_village_row + xls_village_step

        # 统计每月每县总绑定数量
        newsheet.write(xls_village_row, xls_rt_bind_sum + i, binnum_total)


        #更新表格.....
        areabinddata = {
            "area": sheetname, 
            "month": datetmpstr[i], 
            "clientNum": binnum_total , 
            "clientTotal": binnum_total
        }

        #由于201907开始有绑定数，201907不需要计数当月绑定数   
        if datetmpstr[i] != "201907":
            #得出上个月数据
            bind_firstday = BIND_TIME.replace(day=1)
            last_month = bind_firstday - datetime.timedelta(days=1)
            lastmyquery = {
                "area": sheetname,
                "month": last_month.strftime("%Y%m")
            }
            #print(last_month.strftime("%Y%m"))

            lastmydate = area_bindcollection.find_one(lastmyquery)

            if lastmydate:
                areabinddata["clientNum"] = binnum_total - lastmydate["clientTotal"]

        #查询数据库没有数据就插入，有就更新
        myquery = {
            "area": sheetname, 
            "month":datetmpstr[i]
        }       
        mydate = area_bindcollection.find_one(myquery)
        if (mydate == None):
            area_bindcollection.insert(areabinddata)
        else:
            #更新数据库
            mydate["clientTotal"]  = binnum_total
            mydate["clientNum"] =  areabinddata["clientNum"]
            area_bindcollection.update(myquery,mydate)

        i = i + 1
        print(sheetname, BIND_TIME, areabinddata["clientNum"], binnum_total)

    client.close()


if __name__ == "__main__":

    # 读取数据
    book = xlrd.open_workbook(xls_path)
    print("The number of worksheets in %s is %d" % (xls_path, book.nsheets))

    for sheet_names in book.sheet_names():
        print(sheet_names)

    print("now working ........ ")
    newbook = copy(book)

    threads = []

    for i in range(0, book.nsheets):
        sheet = book.sheet_by_index(i)
        newsheet = newbook.get_sheet(i)
        print(book.sheet_names())

        newsheet.col(3).width = 17 * 256
        newsheet.col(4).width = 14 * 256
        newsheet.col(5).width = 15 * 256
        newsheet.col(6).width = 17 * 256
        newsheet.col(7).width = 15 * 256
        newsheet.col(8).width = 15 * 256
        newsheet.col(9).width = 11 * 256

        xls_row_start = 1
        xls_row_end = sheet.nrows
        xls_row_step = 1

        city = CityInfo(sheet.name, xls_row_start, xls_row_end, xls_row_step)
        t = threading.Thread(target=work, name=city.name.encode("utf-8"), args=(city, sheet, newsheet, book.sheet_names()[i]))
        threads.append(t)
        t.start()

    for t in threads:
        t.join()
        print(t)


    try:
        os.remove(xls_new)
    except Exception as e:
        print(e)
        pass

    newbook.save(xls_new)
    print("save  ok")


