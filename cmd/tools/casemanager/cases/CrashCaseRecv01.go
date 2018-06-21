package cases

import (
	"fmt"

	"time"

	"github.com/SmartMeshFoundation/SmartRaiden/cmd/tools/casemanager/models"
	"github.com/SmartMeshFoundation/SmartRaiden/cmd/tools/casemanager/utils"
	"github.com/SmartMeshFoundation/SmartRaiden/params"
)

// CrashCaseRecv01 场景一：ActionInitTargetStateChange
// 收到mtr后崩,它是接收方
// 从节点2向节点6发送45个token，节点6崩后，节点2 锁定45token，节点3锁定45token，转帐失败；重启后，转账继续。
func (cm *CaseManager) CrashCaseRecv01() (err error) {
	env, err := models.NewTestEnv("./cases/CrashCaseRecv01.ENV")
	if err != nil {
		return
	}
	defer env.KillAllRaidenNodes()
	// 源数据
	var transAmount int32
	var msg string
	transAmount = 45
	tokenAddress := env.Tokens[0].Address
	N2, N3, N6 := env.Nodes[0], env.Nodes[1], env.Nodes[2]
	models.Logger.Println(env.CaseName + " BEGIN ====>")
	// 启动节点2,3
	N2.Start(env)
	N3.Start(env)
	// 启动节点6, ActionInitTargetStateChange
	N6.StartWithConditionQuit(env, &params.ConditionQuit{
		QuitEvent: "ActionInitTargetStateChange",
	})

	// 查询节点2，记录cd24数据
	cd23 := utils.GetChannelBetween(N2, N3, tokenAddress)
	cd23.Println("Channel data before transfer send, cd23:")
	// 查询节点3，记录cd36数据
	cd36 := utils.GetChannelBetween(N3, N6, tokenAddress)
	cd36.Println("Channel data before transfer send, cd36:")

	// 节点2向节点6转账20token
	N2.SendTrans(tokenAddress, transAmount, N6.Address, false)
	time.Sleep(time.Second * 3)
	//  崩溃判断
	if N6.IsRunning() {
		panic("Node N6 should be exited,but it still running")
	}

	// 查询cd23，锁定45
	cd23middle := utils.GetChannelBetween(N2, N3, tokenAddress)
	cd23middle.Println("Channel data after transfer send, cd23middle:")
	if cd23middle.LockedAmount != transAmount {
		msg = fmt.Sprintf("Expect locked amount = %d,but got %d ,FAILED!!!", transAmount, cd23middle.PartnerLockedAmount)
		models.Logger.Println(msg)
		return fmt.Errorf(msg)
	}
	// 查询cd36，锁定45
	cd36middle := utils.GetChannelBetween(N3, N6, tokenAddress)
	cd36middle.Println("Channel data after transfer send, cd36middle:")
	if cd36middle.LockedAmount != transAmount {
		msg = fmt.Sprintf("Expect locked amount = %d,but got %d ,FAILED!!!", transAmount, cd36middle.LockedAmount)
		models.Logger.Println(msg)
		return fmt.Errorf(msg)
	}

	// 重启节点6，交易自动继续
	N6.DebugCrash = false
	N6.ConditionQuit = nil
	N6.Name = "RestartNode"
	N6.Start(env)

	// 查询cd23并校验
	cd23new := utils.GetChannelBetween(N2, N3, tokenAddress)
	cd23new.Println("Channel data after transfer success, cd23new:")
	if cd23new.Balance-cd23.Balance != transAmount {
		models.Logger.Println(env.CaseName + " END ====> FAILED")
		return fmt.Errorf("Case [%s] FAILED", env.CaseName)
	}
	// 查询cd36并校验
	cd36new := utils.GetChannelBetween(N3, N6, tokenAddress)
	cd36new.Println("Channel data after transfer success, cd36new:")
	if cd36new.Balance-cd36new.Balance != transAmount {
		models.Logger.Println(env.CaseName + " END ====> FAILED")
		return fmt.Errorf("Case [%s] FAILED", env.CaseName)
	}
	models.Logger.Println(env.CaseName + " END ====> SUCCESS")
	return
}
