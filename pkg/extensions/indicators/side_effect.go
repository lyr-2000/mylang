package indicators

// pureFuncs 是「可安全做纯优化（如公共子表达式消除）」的函数白名单。
//
// 默认安全策略：只有在此白名单里的函数才被视为「无副作用、同输入同输出」，
// 优化器才可对其消除重复计算。其余函数（含用户通过 RegisterFunction /
// RegisterContextFunction 注册的扩展函数）默认视为「有副作用」，不参与纯优化。
//
// 内置指标函数在包 init 时自动登记（它们是库保证的无状态列式函数）；
// 用户扩展函数需显式调用 MarkPure 声明为纯函数后才会被优化。
var pureFuncs = map[string]struct{}{}

// MarkPure 显式声明某函数为纯函数（可优化）。
// 用于用户扩展函数：默认视为有副作用，只有显式声明后才参与 CSE 等纯优化。
// 声明通常应在执行器初始化阶段（单线程）完成。
func MarkPure(name string) {
	pureFuncs[name] = struct{}{}
}

// IsPure 判断函数是否可做纯优化（在纯函数白名单里）。
func IsPure(name string) bool {
	_, ok := pureFuncs[name]
	return ok
}
