package utils

import (
	"github.com/viwii/utils/stack"
)

var (
	Timer_LoopForever int = -1 //无效定时器
	Timer_Stop        int = 0  //定时器运行一次
)

type TimerItem struct {
	TimerFunc func()
}

type ILinerHandle interface {
	OnLinerEvent(interface{})
}

type LinerDispatch struct {
	linerList      *PacketList       //将chan改为发送list,不会再阻塞
	eventHandle    func(interface{}) //事件回调
	needKeep       bool              //是否开始运行
	exit           chan int          //退出
	dataSign       chan int          //数据信号
	isWaiting      bool              //是否正在运行中
	timerIDCounter int64
}

func (ld *LinerDispatch) Init() {
}

func (ld *LinerDispatch) Add(pkt interface{}) {
	ld.linerList.Add(pkt)
	ld.dataSign <- 0
}

func (ld *LinerDispatch) RegiserCallBack(fc func(interface{})) {
	ld.eventHandle = fc
}

func (ld *LinerDispatch) Start() {
	ld.needKeep = true
	go func(l *LinerDispatch) {
		var sendList []interface{}
		l.isWaiting = true
		for l.needKeep {
			select {
			case <-l.exit:
				goto exit
			case <-l.dataSign:
				if l.linerList.Len() > 0 {
					sendList = sendList[0:0]
					packetList := l.linerList.BeginPick()
					sendList = append(sendList, packetList...)
					l.linerList.EndPick()
					if l.eventHandle != nil {
						for _, v := range sendList {
							stack.SafeCall(func(args ...interface{}) error {
								l.eventHandle(args[0])
								return nil
							}, v)

						}
					}
				}
				break
			}
		}
	exit:
		l.isWaiting = false
	}(ld)
}

func (ld *LinerDispatch) Stop() {
	ld.needKeep = false
	ld.exit <- 0
}

func NewLinerDispatch() *LinerDispatch {
	ld := &LinerDispatch{
		linerList: NewPacketList(),
		dataSign:  make(chan int, 100),
	}

	return ld
}
