package ops

import (
	"github.com/wonderivan/logger"
	"os/exec"
	"time"
)

func processBindNumResult(result string) {
	logger.Info("bindnum result: %s", result)
}

//处理逻辑在python脚本里面，go语言只是定时调用pythp脚本
func BindNum() {

	for true {
		cmd := exec.Command("python3", "ops/bindnum/bindnum.py")
		out, err := cmd.CombinedOutput()
		if err != nil {
			logger.Error("cmd failed %s", err.Error())
			logger.Error("cmd output %s", string(out))
		} else {
			processBindNumResult(string(out))
		}
		time.Sleep(24 * time.Hour)
	}
}
