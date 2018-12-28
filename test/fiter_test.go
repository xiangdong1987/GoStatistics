package test

import (
	"GoStatistics/model"
	"GoStatistics/myTool"
	"fmt"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	filter := model.AcTrie{}
	filter.Dictionary = make(map[int32]int)
	json, _ := myTool.ReadAll("/data/go/src/GoStatistics/model/dictionary/dictionary.json")
	var listDic map[string]string
	myTool.Jsondecode(json, &listDic)
	filter.InitDictionary(listDic)
	filter.Root = &model.AcNode{}
	filter.Root.Children = make([]*model.AcNode, filter.DicLength)
	//filter.root.fail=filter.root
	for value, _ := range listDic {
		filter.AddWord(value)
	}
	//初始错误指针
	filter.InitFailPoint()

	t1 := time.Now() // get current time
	result := filter.Match("电动毛贼东的进口量发动机弗兰克的减肥了的房间的路口附近的考虑附近的会计分录进口量将禄口街道可令肌肤考虑对方考虑对方就离开的房间的空间分开了的减肥的考习近平虑荆防颗粒附近的空间分开了的弗兰克的看风景的路口附近的龙卷风两架飞机的健康可令肌肤考虑到进口量的房间的考虑附近六角恐龙减肥的考虑积分来得快接口")
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
	fmt.Println(result)
}
