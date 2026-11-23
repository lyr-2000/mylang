package mylang

// drawFuncs 画图函数名集合。扩展包（如 drawing）在 init 里调用 RegisterDrawFunc 注册，
// 使解析器把对应函数的裸调用识别为「无变量名画图语句」（即使不带修饰符）。
// 仅在包 init 阶段写入，运行期只读，故无并发问题。
var drawFuncs = map[string]struct{}{}

// RegisterDrawFunc 注册一个画图函数名。
func RegisterDrawFunc(name string) {
	drawFuncs[name] = struct{}{}
}

// IsDrawFunc 判断 name 是否为画图函数。
func IsDrawFunc(name string) bool {
	_, ok := drawFuncs[name]
	return ok
}
