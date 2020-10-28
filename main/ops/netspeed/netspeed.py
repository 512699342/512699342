# -*- coding: utf-8-*-

# python 2.7
''' 
>>> pip install xlrd
>>> pip install xlutils
>>> pip install threadpool
>>> pip install paramiko
# >>> pip install enum34
'''
import time,threading
import os
import datetime
import sys
import telnetlib
import xlrd
import re
import logging
import threadpool
import paramiko
import json
# from enum import Enum
from xlutils.copy import copy

import urllib.request
import json
from urllib import parse
import uuid
import http.client
import pandas as pd


mutex=''
logger=''

# 云之讯-短信请求url
sms_yzx_url = "https://open.ucpaas.com/ol/sms/sendsms"
# 云之讯-账号相关信息-应用ID
sms_yzx_appid = "93649dcfefc24dc29afd8f6b2b0a1728"
# 云之讯-账号相关信息-用户sid
sms_yzx_account_Sid = "eb6f0a1b4f3e59ce23903ffb8b2a37c3"
# 云之讯-账号相关信息-密钥
sms_yzx_auth_token = "0278286f8ebecc70d2c0f891cf521892"
# 云之讯-账号相关信息-短信模板
sms_yzx_templateid = "470311"

# 聚合数据-短信平台-请求地址
sms_juhe_url = "http://v.juhe.cn/sms/send"
# 聚合数据-短信平台-短信模板
sms_juhe_templateid = "188710"
# 聚合数据-短信平台-密钥
sms_juhe_auth_token = "403a8fa94b1f8a0a7e71f0f2a13fa410"

# 短信服务平台(0：云之讯短信    1: 聚合短信)
sms_service_choice = 1

#联系人
wenjun_phone_number = '13246806822'
jianbo_phone_number = '15820228727'
weixiang_phone_number = '15989200983'
gengyu_phone_number = '18702016967'

g_diff_value_increase = 50 #增速比较值50M
g_diff_value_decrease = -50 #减速比较值50M
diff_warning_village_count = 30 #速度变化累计超过多少个村上报告警

def juhe_sendsms(appkey, mobile, tpl_id, tpl_value):
    sendurl = sms_juhe_url  # 短信发送的URL,无需修改
    tpl_value = '#content#=' + tpl_value

    #组合参数
    params = {"key":appkey,
              "mobile":mobile,
              "tpl_id":tpl_id,
              "tpl_value":tpl_value
              }

    params = parse.urlencode(params).encode('utf-8')

    wp = urllib.request.urlopen(sendurl, params)
    content = wp.read().decode()  # 获取接口返回内容

    result = json.loads(content)
    # print(result)
    if result:
        error_code = result['error_code']
        if error_code == 0:
            # 发送成功
            smsid = result['result']['sid']
            logger.info("sendsms success,smsid: %s" % (smsid))
        else:
            # 发送失败
            logger.error("sendsms error :(%s) %s" % (error_code, result['reason']))
    else:
        # 请求失败
         logger.error("request sendsms error")

def yzx_sendsms(to, params, temp_id):
    # @param to 手机号码
    # @param params 内容数据 格式为数组 例如：{'12','34'}，如不需替换请填 ''
    # @param temp_id 模板Id
    data = {
        "sid": sms_yzx_account_Sid,
        "token": sms_yzx_auth_token,
        "appid": sms_yzx_appid,
        "templateid": temp_id,
        "param": params,
        "mobile": to,
    }
    # 将字典转换为JSON字符串
    json_data = json.dumps(data)
    # print(json_data)
    # 发送请求头
    headers = {
        'Accept': 'application/json',
        'Content-Type': 'application/json;charset=utf-8',
    }
    connect = http.client.HTTPConnection('open.ucpaas.com')
    # 发送请求
    connect.request(method='POST', url=sms_yzx_url,
                    body=json_data, headers=headers)
    # 获取响应
    resp = connect.getresponse()
    # print(resp)
    # 响应内容
    result = resp.read().decode('utf-8')
    # print(result)
    result = json.loads(result)

    # 发送成功
    # print(result)
    # 如果发送短信成功，返回的字典数据中code字段的值为"000000"
    if result["code"] == "000000":
        # 返回0 表示发送短信成功
        smsid = result['smsid']
        logger.info("sendsms success,smsid: %s ,%s" % (smsid, result['msg']))
        return 0
    else:
        # 返回-1 表示发送失败
        logger.error("sendsms error :(%s), %s" % (result['code'], result['msg']))
        return -1

def send_sms(mobile, value):
    # mobile = '1870xxxx'  # 短信接受者的手机号码
    # value = '#code#=4567'

    if (sms_service_choice == 0):
        yzx_sendsms(mobile, value, sms_yzx_templateid)
    else:
        juhe_sendsms(sms_juhe_auth_token, mobile,
                     sms_juhe_templateid, value)

#初始化日志
def init_logger():
    logger = logging.getLogger("EUHT")
    logger.setLevel(logging.DEBUG)
    logFileName = datetime.datetime.now().strftime("%Y-%m-%d_%H-%M-%S")+".log"
    fh = logging.FileHandler(logFileName)
    fh.setLevel(logging.DEBUG)
    ch = logging.StreamHandler()
    ch.setLevel(logging.DEBUG)
    formatter = logging.Formatter("%(asctime)s-%(name)s-%(levelname)s-%(threadName)s>\t%(message)s")
    ch.setFormatter(formatter)
    fh.setFormatter(formatter)
    #logger.addHandler(ch)
    logger.addHandler(fh)
    return logger

def sshclient_execmd(hostname, port, username, password): 
    try_count = 0
    while  (try_count < 2):
        try:
            paramiko.util.log_to_file("paramiko.log") 

            ssh = paramiko.SSHClient() 
            ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy()) 
            ssh.connect(hostname=hostname, port=port, username=username, password=password)			
            stdin, stdout, stderr = ssh.exec_command ('uptime') 
            stdin.write("Y") # Generally speaking, the first connection, need a simple interaction. 
            uptime_info = stdout.read().decode()    
            logger.info (uptime_info)	
            if(uptime_info.find('up')):
                return ssh		
        except Exception as e:
            try_count = try_count + 1
            logger.error (e)
            logger.error (ssh)
            logger.error ("ssh advantech-Computer fail , errcode = %d  "  % try_count) 	
			
        if (try_count == 2):
            ssh.close()		
            return False   

class Telnet(object):
    __router_login_name = b'admin'
    __router_login_passwd = b'!iK9u5!!fUkjKou!9klIUf!23mjJD123r!87k!'  
    __router_terminal_header = b'>'
    __cap_login_name = b'root\n'
    __zombiestr = "zombie"
    __notexiststr = "not exist"
    __normalstr = "normal"

    router_terminal_enable_header = b'#'
    router_configenable = b'enable\n'
    router_show_version = b'show version\n'
    # def getTelnetSession (self):
    #     return self.session
    @classmethod
    def login_router(cls, ip, port, retry, timeout):
        trycount = 0
        while (trycount < retry):
            try:
                session = telnetlib.Telnet(ip, port, timeout)
                session.set_debuglevel(0) 
                retstr = session.read_until(b"Username: ", 8)
                if(retstr):
                    logger.info("raisecom normal............")				
                else:
                    logger.info("raisecom fault!!!!!!!!!!!!!") 
                    return "R_Fault"
				
                session.write(cls.__router_login_name + b"\n")
                retstr = session.read_until(b"Password: ", 3)
                session.write(cls.__router_login_passwd + b"\n")
                retstr = session.read_until(b">", 5)				
                if (retstr.find(b">") > -1):
                    logger.debug("login router: %s successfully ", ip)
                    return session
                else:
                    logger.error("Router Password error....")
                    return "P_Error"
					
            except Exception as e:
                trycount += 1
                logger.error(e)
                logger.error("telnet router fail , errcode = %d ", trycount) 
            if (trycount == retry):
                break
        return None

    @staticmethod		
    def logout_router(session):
        session.write(b"exit\n")

    @classmethod
    def get_routerarks_info(cls, session, retRouterArk ):
        """ get cap info, including router_count, sn_value, humidity, 
            temperature, channel, txPower, softver, snmpd status.
            :param self: 
            :param session: telnet session of the country router.
            :param retcap: cap obj which save the info.
            :return True: successful
                    False: sometimes write timeout
        """   
        try:
            logger.info("=================== get Router&ARK info ====================")
            session.write(cls.router_configenable)  #进入查看模式
            session.read_until(cls.router_terminal_enable_header,5)						
           
            session.write(cls.router_show_version) 
            readlog = session.read_until(b'Temperature',3)
            retstr = readlog.decode()
            #系统运行时间
            retRouterArk.router_timer = retstr.split("Device uptime   :")[1].split("seconds.")[0] + 'seconds.'		
            logger.info("路由器系统运行时间：" + retRouterArk.router_timer)
			
            #查看路由器版本号
            retRouterArk.router_version = retstr.split("IOS version: ")[1].split("Bootrom version:")[0].split("\n")[0].split()[0]
            logger.info("路由器软件版本：" + retRouterArk.router_version)

            # #查看eth1是否活着。
            # session.write(b'show interface eth1\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth1 is ")[1].split(", Admin status")[0]			
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth1_status = eth_state + '_' + eth_speed
            # logger.info("eth1状态：%s" %retRouterArk.eth1_status)

			# #查看eth2是否活着。
            # session.write(b'show interface eth2\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth2 is ")[1].split(", Admin status")[0]			
		
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth2_status = eth_state + '_' + eth_speed
            # logger.info("eth2状态：%s" %retRouterArk.eth2_status)
	
			# #查看eth3是否活着。
            # session.write(b'show interface eth3\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth3 is ")[1].split(", Admin status")[0]			
		
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth3_status = eth_state + '_' + eth_speed
            # logger.info("eth3状态：%s" %retRouterArk.eth3_status)

			# #查看eth4是否活着。
            # session.write(b'show interface eth4\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth4 is ")[1].split(", Admin status")[0]			
		
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth4_status = eth_state + '_' + eth_speed
            # logger.info("eth4状态：%s" %retRouterArk.eth4_status)

			#查看eth5是否活着。
            session.write(b'show interface eth5\n')
            readlog = session.read_until(b'Media type:',3)
            retstr = readlog.decode()
            eth_state = retstr.split("Interface eth5 is ")[1].split(", Admin status")[0]			
		
            if(eth_state == "up"):
                eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                retRouterArk.eth5_status = eth_state + '_' + eth_speed
            logger.info("eth5状态：%s" %retRouterArk.eth5_status)

            # #查看eth6是否活着。
            # session.write(b'show interface eth6\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth6 is ")[1].split(", Admin status")[0]			
		
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth6_status = eth_state + '_' + eth_speed
            # logger.info("eth6状态：%s" %retRouterArk.eth6_status)
			
			# #查看eth7是否活着。
            # session.write(b'show interface eth7\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth7 is ")[1].split(", Admin status")[0]			
		
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth7_status = eth_state + '_' + eth_speed
            # logger.info("eth7状态：%s" %retRouterArk.eth7_status)

			# #查看eth8是否活着。
            # session.write(b'show interface eth8\n')
            # readlog = session.read_until(b'Media type:',3)
            # retstr = readlog.decode()
            # eth_state = retstr.split("Interface eth8 is ")[1].split(", Admin status")[0]			
		
            # if(eth_state == "up"):
                # eth_speed = retstr.split("current speed is ")[1].split(", duplex")[0]	
                # retRouterArk.eth8_status = eth_state + '_' + eth_speed
            # logger.info("eth8状态：%s" %retRouterArk.eth8_status)
				
		    #ping 村级服务器
            xls_Server_ip_addr = 'ping ' + retRouterArk.ip + '\n'
            xls_Server_ip_addr = xls_Server_ip_addr.encode()
            session.write( xls_Server_ip_addr)
            retstr = session.read_until(b'64 bytes', 10)		
            if(retstr.find(b'64 bytes') > 0):
                logger.info("Ping: %s 村级服务器通....." %retRouterArk.ip)                
                retRouterArk.is_online = "ping_success" 	
            else:
                retRouterArk.is_online = retRouterArk.eth5_status				
                logger.error("Ping: %s 村级服务器不通！！！....." %retRouterArk.ip)	
				
            time.sleep(1)   
            session.write(b'exit\n') # 退出路由器    				
            return True
			
        except Exception as e:
            logger.error(e)
            logger.error("get router info fail")
        return False
			
class Site(object):
    def __init__(self, province, city, country, zoningname, 
                    zoningcode, routerarks, router, rownum):
        self.city = city
        self.routerarks = routerarks
        self.router = router
        self.rownum = rownum
        self.country = country
        self.province = province
        self.zoningname = zoningname
        self.zoningcode = zoningcode

    def dump(self): 
        logger.info("province=%s, city=%s, country=%s, zoningname=%s, "
            "zoningcode=%s, routerip=%s, row=%d, serverip = %s",
            self.province, self.city, self.country, self.zoningname, 
            self.zoningcode, self.router.ip(), self.rownum, self.routerarks.ip)

class Router(object):
    def __init__(self, ip):
        self.__ip = ip
        self.__isOnline = False

    def ip(self):
        return self.__ip

    def is_online(self):
        return self.__isOnline

    def set_status(self, online):
        self.__isOnline = online

class RouterARK(object):
    def __init__ (self, ip):
        self.ip = ip
        self.router_version = ""
        self.ark_status = ""
        self.ark_version = ""
        self.ark_apk = ""
        self.ark_timer = ""
        self.eth1_status = "down"
        self.eth2_status = "down"
        self.eth3_status = "down"
        self.eth4_status = ""
        self.eth5_status = "down"
        self.eth6_status = ""
        self.eth7_status = ""
        self.eth8_status = ""
        self.router_timer = ""		
        self.is_online = ""	
		
    def is_online(self):
        return self.is_online
    
    def dump(self):
        if (GConfig.is_debug):
            logger.info("ip[" + str(self.ip) + "]")
            logger.info("router_version[" + str(self.router_version) + "]")
            logger.info("ark_status[" + str(self.ark_status) + "]")
            logger.info("ark_version[" + str(self.ark_version) + "]")
            logger.info("ark_apk[" + str(self.ark_apk) + "]")
            logger.info("ark_timer[" + str(self.ark_timer) + "]")
            logger.info("eth1_status[" + str(self.eth1_status) + "]")
            logger.info("eth2_status[" + str(self.eth2_status) + "]")
            logger.info("eth3_status[" + str(self.eth3_status) + "]")
            logger.info("eth4_status[" + str(self.eth4_status) + "]")
            logger.info("eth5_status[" + str(self.eth5_status) + "]")
            logger.info("eth6_status[" + str(self.eth6_status) + "]")
            logger.info("eth7_status[" + str(self.eth7_status) + "]")
            logger.info("eth8_status[" + str(self.eth8_status) + "]")
            logger.info("router_timer[" + str(self.router_timer) + "]")
class GConfig(object):
    is_debug = True
    xls_row_step = 1        # 暂时以IP信息表为准，每个站点占4行
    is_use_default_cap_count = True
    default_cap_count = 3   # default count of routerarks
    # select which one to save to newsheet
    is_get_router_version = True
    is_get_ark_status = True
    is_get_ark_version = True
    is_get_ark_apk = True
    is_get_ark_timer = True
    is_get_eth1_status = False
    is_get_eth2_status = False
    is_get_eth3_status = False
    is_get_eth4_status = True
    is_get_eth5_status = True
    is_get_eth6_status = True
    is_get_eth7_status = True
    is_get_eth8_status = True
    is_get_router_timer = True
	
    # write back cols
    col_router_status = 0
    col_router_version = 0
    col_ark_status = 0
    col_ark_version = 0
    col_ark_timer = 0
    col_eth1_status = 0
    col_eth2_status = 0
    col_eth4_status = 0
    col_eth5_status = 0
    col_eth6_status = 0
    col_eth7_status = 0
    col_eth8_status = 0
    col_router_timer = 0
	
    col_end = 0

class TaskResult:
    def __init__(self, router, routerarks, rownum, type, messages, sheethandler):
        self.router = router
        self.routerarks = routerarks
        self.rownum = rownum
        self.type = type
        self.messages = messages
        self.sheethandler = sheethandler

def enum(*sequential, **named):
    enums = dict(zip(sequential, range(len(sequential))), **named)
    return type('Enum', (), enums)

_ColumnNumber = enum(
    Province = 0,
    City = 1,
    Country = 2,
    Zoningname = 3,
    Zoningcode = 4,
    Routerip = 5,
    Channel = 6,
    Capip = 7,
    Serverip = 8
)

_ResultType = enum(
    Ok = 0,
    Router_Offline = 1,
    Routerip_Invalid = 2,
    Cap_Send_Cmd_Fail = 3,
	Router_Error = 4,
)

def checkip(ip):
    p = re.compile('^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$')  
    if (p.match(ip)):  
        return True
    else:
        return False

def work_func(site, sheethandler):
    """ main work of thread, get router status(online/offline) and some params of cap 
        including routerarkstatue, router_count, sn_value, humidity, temperature, 
        channel, txPower, softver, snmpd status
        :param site: target site
        :param sheethandler: handler to save result of new sheet
    """
    logger.info('thread name[%s] id[%d] start...', 
        threading.current_thread().name, threading.current_thread().ident)

    site.dump()
    messages = []
    if ((not site.router.ip()) or (checkip(site.router.ip()) == False)):
        # invalid ip
        logger.error("row[%d] router.ip invalid", site.rownum)
        messages.append("router ip invalid")
        return TaskResult(site.router, site.routerarks, site.rownum, _ResultType.Routerip_Invalid, messages, sheethandler)

    session = Telnet.login_router(site.router.ip(), 2306, 3, 3)
    if (not session ):
        site.router.set_status(False)
        return TaskResult(site.router, site.routerarks, site.rownum, _ResultType.Ok, None, sheethandler)
    elif (session == "R_Fault" or session == "P_Error" ):
        messages.append("router is Fault or Password Error!!!")	
        site.router.set_status(session)
        return TaskResult(site.router, site.routerarks, site.rownum, _ResultType.Router_Error, messages, sheethandler)

    site.router.set_status(True)
    try_get_cap_info_count = 0
    type = _ResultType.Ok
	
    ret = Telnet.get_routerarks_info(session, site.routerarks)	

    site.routerarks.ark_status = site.routerarks.is_online
    if (ret == True and site.routerarks.is_online == "ping_success"):	       
        session_ssh = sshclient_execmd(site.router.ip(), 2205, "nufront", "nufront")		
        #统计村级服务器数据
        if (session_ssh != False):          
            logger.info("远程SSH登陆工控机成功....")	
            stdin, stdout, stderr = session_ssh.exec_command ('uptime') 
            uptime_info = stdout.read().decode()    
            logger.info (uptime_info)	
            if(uptime_info.find('up')):
                site.routerarks.ark_timer = uptime_info[uptime_info.find('up')+3:uptime_info.find('user')-5]				
                logger.info(site.routerarks.ark_timer)		
            
            stdin, stdout, stderr = session_ssh.exec_command ('cat /opt/RouterManager/nuResources/nuDatabase/Config.ini | grep number') 
            ssh_info = stdout.read().decode()
            logger.info (ssh_info)	
            if(ssh_info.find('=')):
               version = ssh_info[ssh_info.find('=')+1:-1]
               logger.info('RouterManager软件版本：%s ' %version)
               site.routerarks.ark_version = version
               #判断VNC进程是否在					
               stdin, stdout, stderr = session_ssh.exec_command ('ls -a /home/nufront/.vnc/') 				
               ssh_info = stdout.read().decode() 
               logger.info (ssh_info)	
               if(ssh_info.find('pid') and ssh_info.find('log')):
                    nufront_pid = 'cat /home/nufront/.vnc/' + ssh_info[ssh_info.find('log')+4:ssh_info.find('pid')+3]
                    logger.info(nufront_pid)
                    stdin, stdout, stderr = session_ssh.exec_command (nufront_pid) 
                    pid_id = stdout.read().decode()
                    logger.info(pid_id)
                    if(len(pid_id)):
                        nufront_pid = 'ps -aux | grep' + pid_id
                        stdin, stdout, stderr = session_ssh.exec_command (nufront_pid) 
                        ssh_info = stdout.read().decode()						
                        if(ssh_info.find(pid_id)):						
                            logger.info('VNC正常，进程号：%s ' %pid_id)
                            site.routerarks.ark_status = "vnc_success"	

            #######################################################################################					
		    #更新speedtest.py
            max_flag = 0
            for i in range(3):
                stdin, stdout, stderr = session_ssh.exec_command ('ls /opt/ | grep speed')
                ssh_info = stdout.read().decode()

                #安装speedtest工具
                if( ssh_info.find('eedtest.py') < 0 ):
                    stdin, stdout, stderr = session_ssh.exec_command ('cd /opt/;wget ftp://euht:euht@59.32.214.147:21/speedtest.py')
                    #time.sleep(7)
                    ssh_info = stdout.read().decode() + stderr.read().decode()
                    logger.info(ssh_info)
                    logger.info(",,,,,,,,,%d " % ssh_info.find('save'))
                    if(ssh_info.find('saved') > -1):
                        logger.info('speedtest工具安装成功...')
                        site.routerarks.ark_apk = "DownloadSuccess"
                    else:
                        logger.info('speedtest工具下载失败...')
                        site.routerarks.ark_apk = "DownloadFail"
                #执行测试带宽
                if(site.routerarks.ark_apk !=  "DownloadFail"):
                    logger.info("speedtest ======> 第%d次开始测速.........",i+1)
                    #stdin, stdout, stderr = session_ssh.exec_command ('python /opt/speedtest.py --server 17251',timeout=120)
                    stdin, stdout, stderr = session_ssh.exec_command ('python /opt/speedtest.py',timeout=120)
                    #time.sleep(80)
                    ssh_info = stdout.read().decode() + stderr.read().decode()
                    logger.info(ssh_info)

                    if(ssh_info.find('ERROR') > 0 or ssh_info.find('timed out') > 0):
                        logger.info("speedtest ======> 第%d次测速失败，再次尝试",i+1)
                    else:
                        logger.info("speedtest ======> 第%d次测速成功.........",i+1)
                        if(max_flag == 0):
                            max_flag = 1
                            site.routerarks.eth6_status = ssh_info.split("Download: ")[1].split(" Mbit")[0]
                            site.routerarks.eth7_status = ssh_info.split("Upload: ")[1].split(" Mbit")[0]
                            site.routerarks.eth8_status = ssh_info.split("]: ")[1].split(" ms")[0]
                            site.routerarks.eth4_status = ssh_info.split("Hosted by ")[1].split(" [")[0]
                        elif ( float(ssh_info.split("Download: ")[1].split(" Mbit")[0]) > float(site.routerarks.eth6_status) ):
                            site.routerarks.eth6_status = ssh_info.split("Download: ")[1].split(" Mbit")[0]
                            site.routerarks.eth7_status = ssh_info.split("Upload: ")[1].split(" Mbit")[0]
                            site.routerarks.eth8_status = ssh_info.split("]: ")[1].split(" ms")[0]
                            site.routerarks.eth4_status = ssh_info.split("Hosted by ")[1].split(" [")[0]
                        #break
                else:
                    site.routerarks.ark_apk = "安装失败"

            if(max_flag == 1):
                site.routerarks.ark_apk = "测速成功"
            else:
                site.routerarks.ark_apk = "测速失败"

            session_ssh.close()                   		 							
        else:
            logger.error("远程SSH登陆工控机失败.............")

    site.routerarks.dump()
    logger.info('thread name[%s] id[%d] end...', threading.current_thread().name, threading.current_thread().ident)
    return TaskResult(site.router, site.routerarks, site.rownum, type, messages, sheethandler)

def sheet_write_back(request, result):
    """ parse the result, then write to the new sheet
        :param request: 
        :param result: the result of the thread performed
    """
    newsheet = result.sheethandler
    row = result.rownum
    col_remarks = GConfig.col_end
    col_site_row = col_remarks + 1
    global mutex
    mutex.acquire()
    is_router_online = result.router.is_online()
    newsheet.write(row, GConfig.col_router_status, is_router_online)

    if (is_router_online):	 
        newsheet.write(row, GConfig.col_router_version, result.routerarks.router_version)
        newsheet.write(row, GConfig.col_ark_status, result.routerarks.ark_status)		 
        newsheet.write(row, GConfig.col_ark_timer, result.routerarks.ark_timer)
        newsheet.write(row, GConfig.col_ark_version, result.routerarks.ark_version)
        newsheet.write(row, GConfig.col_eth5_status, result.routerarks.eth5_status)
        newsheet.write(row, GConfig.col_ark_apk, result.routerarks.ark_apk)	
        newsheet.write(row, GConfig.col_eth6_status, result.routerarks.eth6_status)
        newsheet.write(row, GConfig.col_eth7_status, result.routerarks.eth7_status)				
        newsheet.write(row, GConfig.col_eth8_status, result.routerarks.eth8_status)
        newsheet.write(row, GConfig.col_eth4_status, result.routerarks.eth4_status)	
        newsheet.write(row, GConfig.col_router_timer, result.routerarks.router_timer)
			
        if (result.type != _ResultType.Ok):
            remarks = ""
            for i in range (0, len(result.messages)):
                # write remarks
                remarks += result.messages[i] + ", "
                newsheet.write(row, col_remarks, remarks)
    if (GConfig.is_debug):
        # check accuracy of the target write back row 
        newsheet.write(row, col_site_row, str(row + 1))
    mutex.release() 
    return False

def init_newsheet(srcsheet, newsheet):
    column_offset = 0
    GConfig.col_router_status = srcsheet.ncols
    newsheet.write(0, GConfig.col_router_status, u"路由器状态")
    if (GConfig.is_get_router_version):
        column_offset += 1
        GConfig.col_router_version = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_router_version, u"路由器版本")
		
    for i in range(16, 21):
        newsheet.col(i).width = 256*16

    if (GConfig.is_get_eth5_status):
        column_offset += 1
        GConfig.col_eth5_status = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_eth5_status, u"eth5状态")

    if (GConfig.is_get_ark_status):
        column_offset += 1
        GConfig.col_ark_status = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_ark_status, u"ARK状态")

    if (GConfig.is_get_ark_version):
        column_offset += 1
        GConfig.col_ark_version = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_ark_version, u"ARK版本")

    if (GConfig.is_get_ark_timer):
        column_offset += 1
        GConfig.col_ark_timer = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_ark_timer, u"ARK运行时间")

    if (GConfig.is_get_ark_apk):
        column_offset += 1
        GConfig.col_ark_apk = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_ark_apk, u"测速状态")
		

    if (GConfig.is_get_eth6_status):
        column_offset += 1
        GConfig.col_eth6_status = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_eth6_status, u"下行(Mbps)")

    if (GConfig.is_get_eth7_status):
        column_offset += 1
        GConfig.col_eth7_status = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_eth7_status, u"上行(Mbps)")

    if (GConfig.is_get_eth8_status):
        column_offset += 1
        GConfig.col_eth8_status = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_eth8_status, u"ping(ms)")
	
    if (GConfig.is_get_eth4_status):
        column_offset += 1
        GConfig.col_eth4_status = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_eth4_status, u"服务器")
		
    if (GConfig.is_get_router_timer):
        column_offset += 1
        GConfig.col_router_timer = GConfig.col_router_status + column_offset
        newsheet.write(0, GConfig.col_router_timer, u"路由器运行时间")
		
    GConfig.col_end = GConfig.col_router_status + column_offset + 1
    return

def main():
    global mutex
    logger.info (sys.argv[1])  
    if( sys.argv[1] == '潮州' ):
        logger.info("/************ 潮州市-带宽测速 ************/")
    elif( sys.argv[1] == '韶关' ):
        logger.info("/************ 韶关市-带宽测速 ************/")
    elif( sys.argv[1] == '河源' ):
        logger.info("/************ 河源市-带宽测速 ************/")	
    elif( sys.argv[1] == '梅州' ):
        logger.info("/************ 梅州市-带宽测速 ************/")	
    elif( sys.argv[1] == '广东' ):
        logger.info("/************ 广东省-带宽测速 ************/")	
    elif( sys.argv[1] == '梅江' or sys.argv[1] == '梅县' or sys.argv[1] == '平远' or sys.argv[1] == '蕉岭'\
	or sys.argv[1] == '大埔' or sys.argv[1] == '五华' or sys.argv[1] == '兴宁' or sys.argv[1] == '丰顺'):
        logger.info(sys.argv[1] + " 带宽测速 ************/")
    elif( sys.argv[1] == '龙川' or sys.argv[1] == '和平' or sys.argv[1] == '连平'):
        logger.info(sys.argv[1] + " 带宽测速 ************/")	
    else:
        logger.info("/************ 输入参数有误，请输入潮州 or 韶关 or 河源 or 广东************/")   
        sys.exit(-1)
	
    xls_name = u"./ops/netspeed/acip_gd.xls"
    #xls_path = sys.path[0] + "\\" + xls_name
    xls_path = xls_name
    srcbook = xlrd.open_workbook(xls_path)
    newbook = copy(srcbook)

    logger.info("\nThe number of worksheets in %s is %d", xls_path, srcbook.nsheets)
    for sheet_name in srcbook.sheet_names():
        logger.info(sheet_name)

    mutex = threading.Lock()
    pool = threadpool.ThreadPool(8)
    for i in range (0, srcbook.nsheets):
        #选择地级市执行
        if(sys.argv[1] != "广东" and srcbook.sheet_names()[i].find(sys.argv[1]) < 0):
            continue
		
        sheet = srcbook.sheet_by_index(i)
        newsheet = newbook.get_sheet(i)
        init_newsheet(sheet, newsheet)  # newsheet add target column headers

        #调试用
        #sheet.nrows = 4
        # init tasklist, a site represents a task
        row = 1
        args = []
        while (row < sheet.nrows):
        #while (row < 5):
            #routerarks = ""	
            routerarks = RouterARK(sheet.cell_value(row, _ColumnNumber.Serverip))
            #routerarks.append(routerark)
			
            router = Router(sheet.cell_value(row, _ColumnNumber.Routerip))
            task = Site(sheet.cell_value(row, _ColumnNumber.Province), sheet.cell_value(row, _ColumnNumber.City),
                sheet.cell_value(row, _ColumnNumber.Country), sheet.cell_value(row, _ColumnNumber.Zoningname),
                sheet.cell_value(row, _ColumnNumber.Zoningcode), routerarks, router, row)
            # take site and the current table as parameters, newsheet 
            # use to write back the result after the task was completed     
            arg = [task, newsheet]
            args.append((arg, None))
            row = row + GConfig.xls_row_step
        # submit the task to the thread pool, waiting for the task to be completed
        requests = threadpool.makeRequests(work_func, args, sheet_write_back)
        [pool.putRequest(req) for req in requests]
        time.sleep(1)
        pool.wait()
        logger.info("\ntasks=%d, end\n", len(args))
        time.sleep(3)

    xls_new_path = xls_path.split(".xls")[0] + "_带宽测速_" + sys.argv[1] + datetime.datetime.now().strftime("%Y-%m-%d_%H-%M-%S")+".xls"
    try:
        os.remove(xls_new_path)
    except Exception as e:
        pass
    newbook.save(xls_new_path)
    logger.info("save  ok")

    speed_book = xlrd.open_workbook(xls_new_path)
    
    Column_Country = 3
    ReduceSpeed_Cnt = 0
    sms_string=""
    for i in range (0, speed_book.nsheets):
        #选择地级市执行
        if(sys.argv[1] != "广东" and speed_book.sheet_names()[i].find(sys.argv[1]) < 0):
            continue
        all_speed_site=0
        sheet = speed_book.sheet_by_index(i)

        for row in range(1, sheet.nrows):
            down_speed = sheet.cell_value(row,GConfig.col_eth6_status)
                        
            if(down_speed == ''):
                continue
            else:
                down_speed = float(str(down_speed))
                
            if(down_speed < 10):
                country = sheet.cell_value(row, Column_Country)
                ReduceSpeed_Cnt = ReduceSpeed_Cnt + 1
                logger.info(country + "降速")
            else:
                all_speed_site= all_speed_site+1

    warning_flag = 0
    sms_string, warning_flag = update_upstream_and_handle_diff(xls_new_path)
    logger.info("warning_flag:%d",warning_flag)

    if(ReduceSpeed_Cnt > 0 or warning_flag == 1):
            #sms_string = sms_string +" "+ str(country)+" 总共测速 "+str(all_speed_site)+" 低于10M "+str(ReduceSpeed_Cnt)
        sms_string = sys.argv[1] + "低于10M村有" + str(ReduceSpeed_Cnt) + "个"
        if(ReduceSpeed_Cnt > 0) :
            sms_string = sms_string + sys.argv[1] + "低于10M村有" + str(ReduceSpeed_Cnt) + "个"
        logger.info(sms_string)
        send_sms(weixiang_phone_number, sms_string)
        send_sms(jianbo_phone_number, sms_string)
        send_sms(wenjun_phone_number, sms_string)
        send_sms(gengyu_phone_number, sms_string)


def update_upstream_and_handle_diff(current_xls_path):
    xls_sheet_start = 0
    xls_sheet_end = -1
    update_upstream_enable = 1 #上行带宽数据更新使能，1--开启，-1--关闭
    handle_calculate_diff_value_enable = 1
    handle_statistic_diff_value_enable = 1

    diff_value_increase_count = -1 #增速超过xx统计
    diff_value_decrease_count = -1 #减速超过xx统计

    diff_value_col = -1 #差值所在列
    warning_increase_msg = "" #升速告警信息
    warning_decrease_msg = "" #降速告警信息
    warning_increase_flag = -1 #告警信息标志
    warning_decrease_flag = -1 #告警信息标志
    warning_flag = 0

    xls_upstream_col = 10 #下行流所在列
    xls_upstream_col_name = "" #以日期为表头
    month_now = datetime.datetime.now().month
    day_now = datetime.datetime.now().day

    xls_name_diff = u"./ops/netspeed/acip_gd_带宽测速汇总.xls"#汇总表
    xls_new_path = current_xls_path #"acip_gd_带宽测速_广东2019.xls" #最新下行带宽数据来源表
    book_diff = xlrd.open_workbook(xls_name_diff)
    book_new = xlrd.open_workbook(xls_new_path)
    logger.info("book_new.nsheets:%d",book_new.nsheets)
    xls_new = xls_name_diff #sys.path[0] + "\\" + xls_name_diff[0:len(xls_name_diff) - 4:] + "_" + datetime.datetime.now().strftime('%Y-%m-%d_%H_%M_%S')+".xls"

    if book_diff.nsheets != book_new.nsheets :
        logger.info("Please check sheet num!!!")
        os._exit(0)

    logger.info("The number of worksheets in %s is %d" % (xls_name_diff, book_diff.nsheets))
    logger.info("The number of worksheets in %s is %d" % (xls_new_path, book_new.nsheets))
    #两个book的表名是否一样
    for sheet_names in book_diff.sheet_names():
        logger.info(sheet_names)

    for sheet_names in book_new.sheet_names():
        logger.info(sheet_names)

    if (xls_sheet_end < 0):
        xls_sheet_end = book_diff.nsheets


    writer = pd.ExcelWriter(xls_new)

    for i in range(xls_sheet_start, xls_sheet_end): #轮询各个工作表xls_sheet_end
        #sheet = book.sheet_by_index(i)
        logger.info("开始数据处理：%s,%s",book_diff.sheet_names()[i],book_new.sheet_names()[i])

        df=pd.DataFrame(pd.read_excel(xls_name_diff, book_diff.sheet_names()[i]))
        if update_upstream_enable > 0:
            df_new_upstream = pd.DataFrame(pd.read_excel(xls_new_path, book_new.sheet_names()[i]))
        # print("df.shape:",df.shape[1], df.shape, df_new_upstream.shape) #数据纬度
        xls_upstream_col = df.shape[1]
        xls_upstream_col_name = str(month_now) + "月" + str(day_now) + "日"#取当天日期
        xls_upstream_diff = str(day_now) + "日" + "差值"

        # logger.info(df.index)

        if update_upstream_enable > 0:
            # logger.info("update_upstream_enable > 0")
            df.insert(xls_upstream_col, xls_upstream_col_name,
                    df_new_upstream.iloc[:, 16:16 + 1]) #拷贝上行流数据
            df.insert(xls_upstream_col+1, xls_upstream_diff, 9999) #设置差值表头及默认值

        #计算差值
        if handle_calculate_diff_value_enable > 0:
            # logger.info("handle_calculate_diff_value_enable > 0")
            #轮询该工作表的每行数据
            diff_value_col = xls_upstream_col+1 #差值所在列
            for row in range(0,df.shape[0]):
                diff_value = df.iloc[row:row+1, (xls_upstream_col-0):(xls_upstream_col-0) + 1].values[0] - df.iloc[row:row+1, (xls_upstream_col-1 -1):(xls_upstream_col-1) + 1 -1].values[0]
                # print("差值:", diff_value)
                df.loc[row:row,xls_upstream_col+1:xls_upstream_col+2] = diff_value #设置差值

        #统计差值大于某个数的总数
        if handle_statistic_diff_value_enable > 0:
            # logger.info("handle_statistic_diff_value_enable > 0")
            #轮询该工作表的每行数据
            diff_value_increase_count = 0
            diff_value_decrease_count = 0
            for row in range(0,df.shape[0]):
                # logger.info(df.iloc[row:row+1, diff_value_col:diff_value_col + 1].values[0])
                if df.iloc[row:row+1, diff_value_col:diff_value_col + 1].values[0] > g_diff_value_increase :
                    diff_value_increase_count = diff_value_increase_count + 1

                if df.iloc[row:row+1, diff_value_col:diff_value_col + 1].values[0] < g_diff_value_decrease :
                    diff_value_decrease_count = diff_value_decrease_count + 1

            #预警数量设为df.shape[0]/2，大于该县半数
            if diff_value_increase_count > diff_warning_village_count:#df.shape[0]/4 :
                warning_increase_msg = warning_increase_msg + book_diff.sheet_names()[i] + str(diff_value_increase_count) + "个村 "
                warning_increase_flag = 1
                logger.info(warning_increase_msg)
            if diff_value_decrease_count > diff_warning_village_count:#df.shape[0]/4 :
                warning_decrease_msg = warning_decrease_msg + book_diff.sheet_names()[i] + str(diff_value_decrease_count) + "个村 "
                warning_decrease_flag = 1
                logger.info(warning_decrease_msg)
            logger.info("diff_value increase、decrease:%s,%s", diff_value_increase_count,diff_value_decrease_count)

        df.to_excel(writer, sheet_name=book_diff.sheet_names()[i],index=False, header=True)
        logger.info("处理完成表：%s", book_diff.sheet_names()[i])

    if warning_increase_flag == 1 :
        warning_increase_flag = -1
        warning_flag = 1
        warning_increase_msg  = warning_increase_msg + "网速上升大于" + str(g_diff_value_increase) + "M。"
        logger.info(warning_increase_msg)
    if warning_decrease_flag == 1 :
        warning_decrease_flag = -1
        warning_flag = 1
        warning_decrease_msg  = warning_decrease_msg + "网速下降大于" + str(g_diff_value_decrease) + "M。"
        logger.info(warning_decrease_msg)

    writer.save()
    logger.info("save book：%s", xls_name_diff)
    return (warning_increase_msg + warning_decrease_msg, warning_flag)


if __name__ == "__main__":
    logger = init_logger()
    main()
