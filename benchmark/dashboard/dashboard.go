package dashboard

import (
	"fmt"
	"html/template"
	"log"
	"runtime"
	"time"

	"net/http"

	"github.com/hb-go/conn/pkg/conv"
)

func Index(w http.ResponseWriter, r *http.Request) {
	updateSystemStatus()
	if err := indexTmpl.Execute(w, sysStatus); err != nil {
		log.Print(err)
	}

}

var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>Dashboard</title>
</head>
<body>
<div style="margin-right:auto; margin-left:auto; width:900px;">

<table style="margin-right:auto; margin-left:auto;">
	<tr>
		<th><h2>系统运行状态</h2></th>
	</tr>
	<tr>
		<td>服务运行时间</td>
		<td>{{.Uptime}}</td>
	</tr>
	<tr>
		<td>当前 Goroutines 数量</td>
		<td>{{.NumGoroutine}}</td>
	</tr>
	<tr>
		<td>当前内存使用量</td>
		<td>{{.MemAllocated}}</td>
	</tr>
	<tr>
		<td>所有被分配的内存</td>
		<td>{{.MemTotal}}</td>
	</tr>
	<tr>
		<td>内存占用量</td>
		<td>{{.MemSys}}</td>
	</tr>
	<tr>
		<td>指针查找次数</td>
		<td>{{.Lookups}}</td>
	</tr>
	<tr>
		<td>内存分配次数</td>
		<td>{{.MemMallocs}}</td>
	</tr>
	<tr>
		<td>内存释放次数</td>
		<td>{{.MemFrees}}</td>
	</tr>
	<tr>
		<td>当前 Heap 内存使用量</td>
		<td>{{.HeapAlloc}}</td>
	</tr>
	<tr>
		<td>Heap 内存占用量</td>
		<td>{{.HeapSys}}</td>
	</tr>
	<tr>
		<td>Heap 内存空闲量</td>
		<td>{{.HeapIdle}}</td>
	</tr>
	<tr>
		<td>正在使用的 Heap 内存</td>
		<td>{{.HeapInuse}}</td>
	</tr>
	<tr>
		<td>被释放的 Heap 内存</td>
		<td>{{.HeapReleased}}</td>
	</tr>
	<tr>
		<td>Heap 对象数量</td>
		<td>{{.HeapObjects}}</td>
	</tr>
	<tr>
		<td>启动 Stack 使用量</td>
		<td>{{.StackInuse}}</td>
	</tr>
	<tr>
		<td>被分配的 Stack 内存</td>
		<td>{{.StackSys}}</td>
	</tr>
	<tr>
		<td>MSpan 结构内存使用量</td>
		<td>{{.MSpanInuse}}</td>
	</tr>
	<tr>
		<td>被分配的 MSpan 结构内存</td>
		<td>{{.HeapSys}}</td>
	</tr>
	<tr>
		<td>MCache 结构内存使用量</td>
		<td>{{.MCacheInuse}}</td>
	</tr>
	<tr>
		<td>被分配的 MCache 结构内存</td>
		<td>{{.MCacheSys}}</td>
	</tr>
	<tr>
		<td>被分配的剖析哈希表内存</td>
		<td>{{.BuckHashSys}}</td>
	</tr>
	<tr>
		<td>被分配的 GC 元数据内存</td>
		<td>{{.GCSys}}</td>
	</tr>
	<tr>
		<td>其它被分配的系统内存</td>
		<td>{{.OtherSys}}</td>
	</tr>
	<tr>
		<td>下次 GC 内存回收量</td>
		<td>{{.NextGC}}</td>
	</tr>
	<tr>
		<td>距离上次 GC 时间</td>
		<td>{{.LastGC}}</td>
	</tr>
	<tr>
		<td>GC 暂停时间总量</td>
		<td>{{.PauseTotalNs}}</td>
	</tr>
	<tr>
		<td>上次 GC 暂停时间</td>
		<td>{{.PauseNs}}</td>
	</tr>
	<tr>
		<td>GC 执行次数</td>
		<td>{{.NumGC}}</td>
	</tr>
</table>
<br>
</body>
</html>
`))

var (
	startTime = time.Now()
)

var sysStatus struct {
	Uptime       string
	NumGoroutine int

	// General statistics.
	MemAllocated string // bytes allocated and still in use
	MemTotal     string // bytes allocated (even if freed)
	MemSys       string // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 // number of pointer lookups
	MemMallocs   uint64 // number of mallocs
	MemFrees     uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    string // bytes allocated and still in use
	HeapSys      string // bytes obtained from system
	HeapIdle     string // bytes in idle spans
	HeapInuse    string // bytes in non-idle span
	HeapReleased string // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  string // bootstrap stacks
	StackSys    string
	MSpanInuse  string // mspan structures
	MSpanSys    string
	MCacheInuse string // mcache structures
	MCacheSys   string
	BuckHashSys string // profiling bucket hash table
	GCSys       string // GC metadata
	OtherSys    string // other system allocations

	// Garbage collector statistics.
	NextGC       string // next run in HeapAlloc time (bytes)
	LastGC       string // last run in absolute time (ns)
	PauseTotalNs string
	PauseNs      string // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
	NumGC        uint32
}

func updateSystemStatus() {
	sysStatus.Uptime = conv.TimeSincePro(startTime)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	sysStatus.NumGoroutine = runtime.NumGoroutine()

	sysStatus.MemAllocated = conv.FileSize(int64(m.Alloc))
	sysStatus.MemTotal = conv.FileSize(int64(m.TotalAlloc))
	sysStatus.MemSys = conv.FileSize(int64(m.Sys))
	sysStatus.Lookups = m.Lookups
	sysStatus.MemMallocs = m.Mallocs
	sysStatus.MemFrees = m.Frees

	sysStatus.HeapAlloc = conv.FileSize(int64(m.HeapAlloc))
	sysStatus.HeapSys = conv.FileSize(int64(m.HeapSys))
	sysStatus.HeapIdle = conv.FileSize(int64(m.HeapIdle))
	sysStatus.HeapInuse = conv.FileSize(int64(m.HeapInuse))
	sysStatus.HeapReleased = conv.FileSize(int64(m.HeapReleased))
	sysStatus.HeapObjects = m.HeapObjects

	sysStatus.StackInuse = conv.FileSize(int64(m.StackInuse))
	sysStatus.StackSys = conv.FileSize(int64(m.StackSys))
	sysStatus.MSpanInuse = conv.FileSize(int64(m.MSpanInuse))
	sysStatus.MSpanSys = conv.FileSize(int64(m.MSpanSys))
	sysStatus.MCacheInuse = conv.FileSize(int64(m.MCacheInuse))
	sysStatus.MCacheSys = conv.FileSize(int64(m.MCacheSys))
	sysStatus.BuckHashSys = conv.FileSize(int64(m.BuckHashSys))
	sysStatus.GCSys = conv.FileSize(int64(m.GCSys))
	sysStatus.OtherSys = conv.FileSize(int64(m.OtherSys))

	sysStatus.NextGC = conv.FileSize(int64(m.NextGC))
	sysStatus.LastGC = fmt.Sprintf("%.3fs", float64(time.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000)
	sysStatus.PauseTotalNs = fmt.Sprintf("%.3fs", float64(m.PauseTotalNs)/1000/1000/1000)
	sysStatus.PauseNs = fmt.Sprintf("%.6fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000)
	sysStatus.NumGC = m.NumGC
}
