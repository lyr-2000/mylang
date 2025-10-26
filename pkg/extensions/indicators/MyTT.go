package indicators

import (
	"fmt"
	"math"
	"reflect"
)

// Series 序列类型，用于表示时间序列数据
type Series []float64

// NewSeries 创建新序列
func NewSeries(data []float64) Series {
	return Series(data)
}

// Len 返回序列长度
func (s Series) Len() int {
	return len(s)
}

// At 获取指定位置的元素
func (s Series) At(i int) float64 {
	if i < 0 || i >= len(s) {
		return math.NaN()
	}
	return s[i]
}

// Last 获取最后一个元素
func (s Series) Last() float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	return s[len(s)-1]
}

// Slice 获取子序列
func (s Series) Slice(start, end int) Series {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start >= end {
		return Series{}
	}
	return s[start:end]
}

// ------------------ 0级：核心工具函数 --------------------------------------------

// RD 四舍五入取D位小数
func RD(N float64, D int) float64 {
	multiplier := math.Pow(10, float64(D))
	return math.Round(N*multiplier) / multiplier
}

// RDS 对序列进行四舍五入取D位小数
func RDS(S Series, D int) Series {
	result := make(Series, len(S))
	multiplier := math.Pow(10, float64(D))
	for i, v := range S {
		result[i] = math.Round(v*multiplier) / multiplier
	}
	return result
}

// RET 返回序列倒数第N个值，默认返回最后一个
func RET(S Series, N int) float64 {
	if len(S) == 0 {
		return math.NaN()
	}
	if N <= 0 {
		N = 1
	}
	if N > len(S) {
		return math.NaN()
	}
	return S[len(S)-N]
}

// ABS 返回绝对值
func ABS(S Series) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = math.Abs(v)
	}
	return result
}

// LN 求自然对数
func LN(S Series) Series {
	result := make(Series, len(S))
	for i, v := range S {
		if v <= 0 {
			result[i] = math.NaN()
		} else {
			result[i] = math.Log(v)
		}
	}
	return result
}

// POW 求S的N次方
func POW(S Series, N float64) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = math.Pow(v, N)
	}
	return result
}

// SQRT 求平方根
func SQRT(S Series) Series {
	result := make(Series, len(S))
	for i, v := range S {
		if v < 0 {
			result[i] = math.NaN()
		} else {
			result[i] = math.Sqrt(v)
		}
	}
	return result
}

// SIN 求正弦值（弧度）
func SIN(S Series) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = math.Sin(v)
	}
	return result
}

// COS 求余弦值（弧度）
func COS(S Series) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = math.Cos(v)
	}
	return result
}

// TAN 求正切值（弧度）
func TAN(S Series) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = math.Tan(v)
	}
	return result
}

// MAX 序列最大值
func MAX(S1, S2 Series) Series {
	result := make(Series, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = math.Max(S1[i], S2[i])
		} else {
			result[i] = S1[i]
		}
	}
	return result
}

// MAXS 序列与标量的最大值（支持广播）
func MAXS(S Series, scalar float64) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = math.Max(v, scalar)
	}
	return result
}

// MIN 序列最小值
func MIN(S1, S2 Series) Series {
	result := make(Series, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = math.Min(S1[i], S2[i])
		} else {
			result[i] = S1[i]
		}
	}
	return result
}

// IF 序列布尔判断
func IF(S []bool, A, B Series) Series {
	result := make(Series, len(S))
	for i, condition := range S {
		if condition && i < len(A) {
			result[i] = A[i]
		} else if i < len(B) {
			result[i] = B[i]
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// REF 对序列整体下移动N，返回序列
func REF(S Series, N int) Series {
	if N <= 0 {
		return S
	}
	result := make(Series, len(S))
	for i := range S {
		if i >= N {
			result[i] = S[i-N]
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// DIFF 前一个值减后一个值
func DIFF(S Series, N int) Series {
	if N <= 0 {
		N = 1
	}
	result := make(Series, len(S))
	for i := range S {
		if i >= N {
			result[i] = S[i] - S[i-N]
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// STD 求序列的N日标准差
func STD(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			slice := S[i-N+1 : i+1]
			mean := 0.0
			for _, v := range slice {
				if !math.IsNaN(v) {
					mean += v
				}
			}
			mean /= float64(len(slice))

			variance := 0.0
			for _, v := range slice {
				if !math.IsNaN(v) {
					variance += math.Pow(v-mean, 2)
				}
			}
			variance /= float64(len(slice))
			result[i] = math.Sqrt(variance)
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// SUM 对序列求N天累计和
func SUM(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if N > 0 {
			if i >= N-1 {
				sum := 0.0
				for j := i - N + 1; j <= i; j++ {
					if !math.IsNaN(S[j]) {
						sum += S[j]
					}
				}
				result[i] = sum
			} else {
				result[i] = math.NaN()
			}
		} else {
			// N=0时对序列所有依次求和
			sum := 0.0
			for j := 0; j <= i; j++ {
				if !math.IsNaN(S[j]) {
					sum += S[j]
				}
			}
			result[i] = sum
		}
	}
	return result
}

// MA 求序列的N日简单移动平均值
func MA(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			sum := 0.0
			for j := i - N + 1; j <= i; j++ {
				if !math.IsNaN(S[j]) {
					sum += S[j]
				}
			}
			result[i] = sum / float64(N)
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// EMA 指数移动平均
func EMA(S Series, N int) Series {
	if N <= 0 {
		return S
	}
	alpha := 2.0 / float64(N+1)
	result := make(Series, len(S))

	// 找到第一个非NaN值作为初始值
	var firstValid float64
	var firstIndex int
	for i, v := range S {
		if !math.IsNaN(v) {
			firstValid = v
			firstIndex = i
			break
		}
	}

	if math.IsNaN(firstValid) {
		return result // 全部都是NaN
	}

	result[firstIndex] = firstValid

	for i := firstIndex + 1; i < len(S); i++ {
		if !math.IsNaN(S[i]) {
			result[i] = alpha*S[i] + (1-alpha)*result[i-1]
		} else {
			result[i] = result[i-1]
		}
	}
	return result
}

// SMA 中国式的SMA
func SMA(S Series, N int, M float64) Series {
	if N <= 0 {
		return S
	}
	alpha := M / float64(N)
	result := make(Series, len(S))

	// 找到第一个非NaN值作为初始值
	var firstValid float64
	var firstIndex int
	for i, v := range S {
		if !math.IsNaN(v) {
			firstValid = v
			firstIndex = i
			break
		}
	}

	if math.IsNaN(firstValid) {
		return result // 全部都是NaN
	}

	result[firstIndex] = firstValid

	for i := firstIndex + 1; i < len(S); i++ {
		if !math.IsNaN(S[i]) {
			result[i] = alpha*S[i] + (1-alpha)*result[i-1]
		} else {
			result[i] = result[i-1]
		}
	}
	return result
}

// WMA 加权移动平均
func WMA(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			sum := 0.0
			weightSum := 0.0
			for j := 0; j < N; j++ {
				weight := float64(j + 1)
				if !math.IsNaN(S[i-N+1+j]) {
					sum += S[i-N+1+j] * weight
					weightSum += weight
				}
			}
			if weightSum > 0 {
				result[i] = sum / weightSum
			} else {
				result[i] = math.NaN()
			}
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// DMA 动态移动平均
func DMA(S Series, A float64) Series {
	if A <= 0 || A >= 1 {
		return S
	}
	result := make(Series, len(S))

	// 找到第一个非NaN值作为初始值
	var firstValid float64
	var firstIndex int
	for i, v := range S {
		if !math.IsNaN(v) {
			firstValid = v
			firstIndex = i
			break
		}
	}

	if math.IsNaN(firstValid) {
		return result // 全部都是NaN
	}

	result[firstIndex] = firstValid

	for i := firstIndex + 1; i < len(S); i++ {
		if !math.IsNaN(S[i]) {
			result[i] = A*S[i] + (1-A)*result[i-1]
		} else {
			result[i] = result[i-1]
		}
	}
	return result
}

// HHV 最近N天最高价
func HHV(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			max := S[i-N+1]
			for j := i - N + 2; j <= i; j++ {
				if !math.IsNaN(S[j]) && S[j] > max {
					max = S[j]
				}
			}
			result[i] = max
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// LLV 最近N天最低价
func LLV(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			min := S[i-N+1]
			for j := i - N + 2; j <= i; j++ {
				if !math.IsNaN(S[j]) && S[j] < min {
					min = S[j]
				}
			}
			result[i] = min
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// ------------------ 1级：应用层函数 ---------------------------------

// COUNT 最近N天满足条件的天数
func COUNT(S []bool, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			count := 0
			for j := i - N + 1; j <= i; j++ {
				if j >= 0 && j < len(S) && S[j] {
					count++
				}
			}
			result[i] = float64(count)
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// EVERY 最近N天是否都是True
func EVERY(S []bool, N int) []bool {
	result := make([]bool, len(S))
	for i := range S {
		if i >= N-1 {
			allTrue := true
			for j := i - N + 1; j <= i; j++ {
				if j < 0 || j >= len(S) || !S[j] {
					allTrue = false
					break
				}
			}
			result[i] = allTrue
		} else {
			result[i] = false
		}
	}
	return result
}

// EXIST N日内是否存在一天满足条件
func EXIST(S []bool, N int) []bool {
	result := make([]bool, len(S))
	for i := range S {
		if i >= N-1 {
			exists := false
			for j := i - N + 1; j <= i; j++ {
				if j >= 0 && j < len(S) && S[j] {
					exists = true
					break
				}
			}
			result[i] = exists
		} else {
			result[i] = false
		}
	}
	return result
}

// CROSS 判断向上金叉穿越
func CROSS(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := 1; i < len(S1); i++ {
		if i < len(S2) {
			result[i] = !(S1[i-1] > S2[i-1]) && (S1[i] > S2[i])
		}
	}
	return result
}

// ------------------ 2级：技术指标函数 ---------------------------------

// MACD MACD指标
func MACD(CLOSE Series, SHORT, LONG, M int) (Series, Series, Series) {
	DIF := SUB(EMA(CLOSE, SHORT), EMA(CLOSE, LONG))
	DEA := EMA(DIF, M)
	MACD := MULS(SUB(DIF, DEA), 2)
	return DIF, DEA, MACD
}

// KDJ KDJ指标
func KDJ(CLOSE, HIGH, LOW Series, N, M1, M2 int) (Series, Series, Series) {
	RSV := DIV(MULS(SUB(CLOSE, LLV(LOW, N)), 100), SUB(HHV(HIGH, N), LLV(LOW, N)))
	K := EMA(RSV, M1*2-1)
	D := EMA(K, M2*2-1)
	J := SUB(MULS(K, 3), MULS(D, 2))
	return K, D, J
}

// RSI RSI指标
func RSI(CLOSE Series, N int) Series {
	DIF := SUB(CLOSE, REF(CLOSE, 1))
	// 使用MAXS函数处理标量0的广播，对应Python的MAX(DIF, 0)
	rs := DIV(SMA(MAXS(DIF, 0), N, 1), SMA(ABS(DIF), N, 1))
	result := MULS(rs, 100)
	
	// 应用四舍五入到2位小数（对应Python的RD函数）
	return RDS(result, 2)
}

// BOLL 布林带指标
func BOLL(CLOSE Series, N int, P float64) (Series, Series, Series) {
	MID := MA(CLOSE, N)
	UPPER := ADD(MID, MULS(STD(CLOSE, N), P))
	LOWER := SUB(MID, MULS(STD(CLOSE, N), P))
	return UPPER, MID, LOWER
}

// WR W&R威廉指标
func WR(CLOSE, HIGH, LOW Series, N, N1 int) (Series, Series) {
	WR := DIV(MULS(SUB(HHV(HIGH, N), CLOSE), 100), SUB(HHV(HIGH, N), LLV(LOW, N)))
	WR1 := DIV(MULS(SUB(HHV(HIGH, N1), CLOSE), 100), SUB(HHV(HIGH, N1), LLV(LOW, N1)))
	return WR, WR1
}

// BIAS 乖离率
func BIAS(CLOSE Series, L1, L2, L3 int) (Series, Series, Series) {
	BIAS1 := DIV(MULS(SUB(CLOSE, MA(CLOSE, L1)), 100), MA(CLOSE, L1))
	BIAS2 := DIV(MULS(SUB(CLOSE, MA(CLOSE, L2)), 100), MA(CLOSE, L2))
	BIAS3 := DIV(MULS(SUB(CLOSE, MA(CLOSE, L3)), 100), MA(CLOSE, L3))
	return BIAS1, BIAS2, BIAS3
}

// PSY 心理线指标
func PSY(CLOSE Series, N, M int) (Series, Series) {
	PSY := DIVS(MULS(COUNT(GreaterThan(CLOSE, REF(CLOSE, 1)), N), 100), float64(N))
	PSYMA := MA(PSY, M)
	return PSY, PSYMA
}

// CCI 顺势指标
func CCI(CLOSE, HIGH, LOW Series, N int) Series {
	TP := DIVS(ADD(ADD(HIGH, LOW), CLOSE), 3)
	return DIV(SUB(TP, MA(TP, N)), MULS(AVEDEV(TP, N), 0.015))
}

// ATR 真实波动N日平均值
func ATR(CLOSE, HIGH, LOW Series, N int) Series {
	TR := MAX(MAX(SUB(HIGH, LOW), ABS(SUB(REF(CLOSE, 1), HIGH))), ABS(SUB(REF(CLOSE, 1), LOW)))
	return MA(TR, N)
}

// BBI 多空指标
func BBI(CLOSE Series, M1, M2, M3, M4 int) Series {
	return DIVS(ADD(ADD(ADD(MA(CLOSE, M1), MA(CLOSE, M2)), MA(CLOSE, M3)), MA(CLOSE, M4)), 4)
}

// DMI 动向指标
func DMI(CLOSE, HIGH, LOW Series, M1, M2 int) (Series, Series, Series, Series) {
	TR := SUM(MAX(MAX(SUB(HIGH, LOW), ABS(SUB(HIGH, REF(CLOSE, 1)))), ABS(SUB(LOW, REF(CLOSE, 1)))), M1)
	HD := SUB(HIGH, REF(HIGH, 1))
	LD := SUB(REF(LOW, 1), LOW)
	DMP := SUM(IF(GreaterThan(HD, NewSeries([]float64{0})), HD, NewSeries([]float64{0})), M1)
	DMM := SUM(IF(GreaterThan(LD, NewSeries([]float64{0})), LD, NewSeries([]float64{0})), M1)
	PDI := DIV(MULS(DMP, 100), TR)
	MDI := DIV(MULS(DMM, 100), TR)
	ADX := MA(ABS(DIV(SUB(MDI, PDI), ADD(PDI, MDI))), M2)
	ADXR := DIVS(ADD(ADX, REF(ADX, M2)), 2)
	return PDI, MDI, ADX, ADXR
}

// TAQ 唐安奇通道(海龟)交易指标
func TAQ(HIGH, LOW Series, N int) (Series, Series, Series) {
	UP := HHV(HIGH, N)
	DOWN := LLV(LOW, N)
	MID := DIVS(ADD(UP, DOWN), 2)
	return UP, MID, DOWN
}

// KTN 肯特纳交易通道
func KTN(CLOSE, HIGH, LOW Series, N, M int) (Series, Series, Series) {
	MID := EMA(DIVS(ADD(ADD(HIGH, LOW), CLOSE), 3), N)
	ATRN := ATR(CLOSE, HIGH, LOW, M)
	UPPER := ADD(MID, MULS(ATRN, 2))
	LOWER := SUB(MID, MULS(ATRN, 2))
	return UPPER, MID, LOWER
}

// TRIX 三重指数平滑平均线
func TRIX(CLOSE Series, M1, M2 int) (Series, Series) {
	TR := EMA(EMA(EMA(CLOSE, M1), M1), M1)
	TRIX := DIV(MULS(SUB(TR, REF(TR, 1)), 100), REF(TR, 1))
	TRMA := MA(TRIX, M2)
	return TRIX, TRMA
}

// VR 容量比率
func VR(CLOSE, VOL Series, M1 int) Series {
	LC := REF(CLOSE, 1)
	return DIV(MULS(SUM(IF(GreaterThan(CLOSE, LC), VOL, NewSeries([]float64{0})), M1), 100), SUM(IF(LessThanOrEqual(CLOSE, LC), VOL, NewSeries([]float64{0})), M1))
}

// CR 价格动量指标
func CR(CLOSE, HIGH, LOW Series, N int) Series {
	MID := DIVS(REF(ADD(ADD(HIGH, LOW), CLOSE), 1), 3)
	return DIV(MULS(SUM(MAX(SUB(HIGH, MID), NewSeries([]float64{0})), N), 100), SUM(MAX(SUB(MID, LOW), NewSeries([]float64{0})), N))
}

// EMV 简易波动指标
func EMV(HIGH, LOW, VOL Series, N, M int) (Series, Series) {
	VOLUME := DIV(MA(VOL, N), VOL)
	MID := DIV(MULS(SUB(ADD(HIGH, LOW), REF(ADD(HIGH, LOW), 1)), 100), ADD(HIGH, LOW))
	EMV := MA(DIV(MUL(MUL(MID, VOLUME), SUB(HIGH, LOW)), MA(SUB(HIGH, LOW), N)), N)
	MAEMV := MA(EMV, M)
	return EMV, MAEMV
}

// DPO 区间震荡线
func DPO(CLOSE Series, M1, M2, M3 int) (Series, Series) {
	DPO := SUB(CLOSE, REF(MA(CLOSE, M1), M2))
	MADPO := MA(DPO, M3)
	return DPO, MADPO
}

// BRAR BRAR-ARBR情绪指标
func BRAR(OPEN, CLOSE, HIGH, LOW Series, M1 int) (Series, Series) {
	AR := DIV(MULS(SUM(SUB(HIGH, OPEN), M1), 100), SUM(SUB(OPEN, LOW), M1))
	BR := DIV(MULS(SUM(MAX(SUB(HIGH, REF(CLOSE, 1)), NewSeries([]float64{0})), M1), 100), SUM(MAX(SUB(REF(CLOSE, 1), LOW), NewSeries([]float64{0})), M1))
	return AR, BR
}

// DFMA 平行线差指标
func DFMA(CLOSE Series, N1, N2, M int) (Series, Series) {
	DIF := SUB(MA(CLOSE, N1), MA(CLOSE, N2))
	DIFMA := MA(DIF, M)
	return DIF, DIFMA
}

// MTM 动量指标
func MTM(CLOSE Series, N, M int) (Series, Series) {
	MTM := SUB(CLOSE, REF(CLOSE, N))
	MTMMA := MA(MTM, M)
	return MTM, MTMMA
}

// MASS 梅斯线
func MASS(HIGH, LOW Series, N1, N2, M int) (Series, Series) {
	MASS := SUM(DIV(MA(SUB(HIGH, LOW), N1), MA(MA(SUB(HIGH, LOW), N1), N1)), N2)
	MA_MASS := MA(MASS, M)
	return MASS, MA_MASS
}

// ROC 变动率指标
func ROC(CLOSE Series, N, M int) (Series, Series) {
	ROC := DIV(MULS(SUB(CLOSE, REF(CLOSE, N)), 100), REF(CLOSE, N))
	MAROC := MA(ROC, M)
	return ROC, MAROC
}

// EXPMA EMA指数平均数指标
func EXPMA(CLOSE Series, N1, N2 int) (Series, Series) {
	return EMA(CLOSE, N1), EMA(CLOSE, N2)
}

// OBV 能量潮指标
func OBV(CLOSE, VOL Series) Series {
	return DIVS(SUM(IF(GreaterThan(CLOSE, REF(CLOSE, 1)), VOL, IF(LessThan(CLOSE, REF(CLOSE, 1)), MULS(VOL, -1), NewSeries([]float64{0}))), 0), 10000)
}

// MFI MFI指标是成交量的RSI指标
func MFI(CLOSE, HIGH, LOW, VOL Series, N int) Series {
	TYP := DIVS(ADD(ADD(HIGH, LOW), CLOSE), 3)
	V1 := DIV(SUM(IF(GreaterThan(TYP, REF(TYP, 1)), MUL(TYP, VOL), NewSeries([]float64{0})), N), SUM(IF(LessThan(TYP, REF(TYP, 1)), MUL(TYP, VOL), NewSeries([]float64{0})), N))
	return SUB(NewSeries([]float64{100}), DIVS(NewSeries([]float64{100}), 1+V1[0]))
}

// ASI 振动升降指标
func ASI(OPEN, CLOSE, HIGH, LOW Series, M1, M2 int) (Series, Series) {
	LC := REF(CLOSE, 1)
	AA := ABS(SUB(HIGH, LC))
	BB := ABS(SUB(LOW, LC))
	CC := ABS(SUB(HIGH, REF(LOW, 1)))
	DD := ABS(SUB(LC, REF(OPEN, 1)))

	R := IF(GreaterThan(AA, BB), ADD(ADD(AA, DIVS(BB, 2)), DIVS(DD, 4)), IF(GreaterThan(BB, CC), ADD(ADD(BB, DIVS(AA, 2)), DIVS(DD, 4)), ADD(CC, DIVS(DD, 4))))

	X := ADD(SUB(CLOSE, LC), ADD(DIVS(SUB(CLOSE, OPEN), 2), SUB(LC, REF(OPEN, 1))))
	SI := DIV(MUL(MULS(X, 16), MAX(AA, BB)), R)
	ASI := SUM(SI, M1)
	ASIT := MA(ASI, M2)
	return ASI, ASIT
}

// XSII 薛斯通道II
func XSII(CLOSE, HIGH, LOW Series, N int, M float64) (Series, Series, Series, Series) {
	AA := MA(DIVS(ADD(ADD(MULS(CLOSE, 2), HIGH), LOW), 4), 5)
	TD1 := DIVS(MULS(AA, float64(N)), 100)
	TD2 := DIVS(MULS(AA, 200-float64(N)), 100)
	DD := MA(CLOSE, 20) // 简化处理，使用MA替代DMA
	TD3 := MULS(DD, 1+M/100)
	TD4 := MULS(DD, 1-M/100)
	return TD1, TD2, TD3, TD4
}

// 辅助函数：序列运算

// ADD 序列加法
func ADD(S1, S2 Series) Series {
	result := make(Series, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] + S2[i]
		} else {
			result[i] = S1[i]
		}
	}
	return result
}

// SUB 序列减法
func SUB(S1, S2 Series) Series {
	result := make(Series, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] - S2[i]
		} else {
			result[i] = S1[i]
		}
	}
	return result
}

// MUL 序列乘法
func MUL(S1, S2 Series) Series {
	result := make(Series, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] * S2[i]
		} else {
			result[i] = S1[i]
		}
	}
	return result
}

// MULS 序列与标量乘法
func MULS(S Series, scalar float64) Series {
	result := make(Series, len(S))
	for i, v := range S {
		result[i] = v * scalar
	}
	return result
}

// DIV 序列除法
func DIV(S1, S2 Series) Series {
	result := make(Series, len(S1))
	for i := range S1 {
		if i < len(S2) && S2[i] != 0 {
			result[i] = S1[i] / S2[i]
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// DIVS 序列与标量除法
func DIVS(S Series, scalar float64) Series {
	result := make(Series, len(S))
	for i, v := range S {
		if scalar != 0 {
			result[i] = v / scalar
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// AVEDEV 平均绝对偏差
func AVEDEV(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			slice := S[i-N+1 : i+1]
			mean := 0.0
			for _, v := range slice {
				if !math.IsNaN(v) {
					mean += v
				}
			}
			mean /= float64(len(slice))

			deviation := 0.0
			for _, v := range slice {
				if !math.IsNaN(v) {
					deviation += math.Abs(v - mean)
				}
			}
			result[i] = deviation / float64(len(slice))
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// GreaterThan 序列比较：S1 > S2
func GreaterThan(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] > S2[i]
		} else {
			result[i] = false
		}
	}
	return result
}

// LessThan 序列比较：S1 < S2
func LessThan(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] < S2[i]
		} else {
			result[i] = false
		}
	}
	return result
}

// LessThanOrEqual 序列比较：S1 <= S2
func LessThanOrEqual(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] <= S2[i]
		} else {
			result[i] = false
		}
	}
	return result
}

// Equal 序列比较：S1 == S2
func Equal(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] == S2[i]
		} else {
			result[i] = false
		}
	}
	return result
}

// NotEqual 序列比较：S1 != S2
func NotEqual(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] != S2[i]
		} else {
			result[i] = true
		}
	}
	return result
}

// GreaterThanOrEqual 序列比较：S1 >= S2
func GreaterThanOrEqual(S1, S2 Series) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i < len(S2) {
			result[i] = S1[i] >= S2[i]
		} else {
			result[i] = true
		}
	}
	return result
}

// SLOPE 返回S序列N周期回线性回归斜率
func SLOPE(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			slice := S[i-N+1 : i+1]
			if len(slice) < 2 {
				result[i] = math.NaN()
				continue
			}

			// 计算线性回归斜率
			sumX := 0.0
			sumY := 0.0
			sumXY := 0.0
			sumXX := 0.0

			for j, v := range slice {
				if !math.IsNaN(v) {
					x := float64(j)
					y := v
					sumX += x
					sumY += y
					sumXY += x * y
					sumXX += x * x
				}
			}

			n := float64(len(slice))
			slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
			result[i] = slope
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// FORCAST 返回S序列N周期回线性回归后的预测值
func FORCAST(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			slice := S[i-N+1 : i+1]
			if len(slice) < 2 {
				result[i] = math.NaN()
				continue
			}

			// 计算线性回归
			sumX := 0.0
			sumY := 0.0
			sumXY := 0.0
			sumXX := 0.0

			for j, v := range slice {
				if !math.IsNaN(v) {
					x := float64(j)
					y := v
					sumX += x
					sumY += y
					sumXY += x * y
					sumXX += x * x
				}
			}

			n := float64(len(slice))
			slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
			intercept := (sumY - slope*sumX) / n

			// 预测下一个值
			predictX := float64(N - 1)
			predictY := slope*predictX + intercept
			result[i] = predictY
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// BARSLAST 上一次条件成立到当前的周期
func BARSLAST(S []bool) Series {
	result := make(Series, len(S))
	lastTrueIndex := -1

	for i, condition := range S {
		if condition {
			lastTrueIndex = i
			result[i] = 0
		} else {
			if lastTrueIndex >= 0 {
				result[i] = float64(i - lastTrueIndex)
			} else {
				result[i] = math.NaN()
			}
		}
	}
	return result
}

// VALUEWHEN 当S条件成立时，取X的当前值，否则取VALUEWHEN的上个成立时的X值
func VALUEWHEN(S []bool, X Series) Series {
	result := make(Series, len(S))
	var lastValue float64 = math.NaN()

	for i, condition := range S {
		if condition && i < len(X) {
			lastValue = X[i]
			result[i] = lastValue
		} else {
			result[i] = lastValue
		}
	}
	return result
}

// BETWEEN S处于A和B之间时为真
func BETWEEN(S, A, B Series) []bool {
	result := make([]bool, len(S))
	for i := range S {
		if i < len(A) && i < len(B) {
			result[i] = ((A[i] < S[i]) && (S[i] < B[i])) || ((A[i] > S[i]) && (S[i] > B[i]))
		} else {
			result[i] = false
		}
	}
	return result
}

// LONGCROSS 两条线维持一定周期后交叉
func LONGCROSS(S1, S2 Series, N int) []bool {
	result := make([]bool, len(S1))
	for i := range S1 {
		if i >= N && i < len(S2) {
			// 检查前N个周期内S1是否都小于S2
			allLess := true
			for j := i - N + 1; j < i; j++ {
				if j >= 0 && j < len(S1) && j < len(S2) {
					if S1[j] >= S2[j] {
						allLess = false
						break
					}
				}
			}
			// 当前周期S1是否大于S2
			currentCross := S1[i] > S2[i]
			result[i] = allLess && currentCross
		} else {
			result[i] = false
		}
	}
	return result
}

// FILTER 函数，S满足条件后，将其后N周期内的数据置为0
func FILTER(S []bool, N int) []bool {
	result := make([]bool, len(S))
	copy(result, S)

	for i, condition := range S {
		if condition {
			// 将后续N个周期置为false
			for j := i + 1; j < len(result) && j < i+1+N; j++ {
				result[j] = false
			}
		}
	}
	return result
}

// LAST 从前A日到前B日一直满足S_BOOL条件
func LAST(S []bool, A, B int) []bool {
	result := make([]bool, len(S))
	for i := range S {
		if i >= A {
			allTrue := true
			for j := i - A + B; j <= i-B; j++ {
				if j < 0 || j >= len(S) || !S[j] {
					allTrue = false
					break
				}
			}
			result[i] = allTrue
		} else {
			result[i] = false
		}
	}
	return result
}

// BARSSINCEN N周期内第一次S条件成立到现在的周期数
func BARSSINCEN(S []bool, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			found := false
			for j := i - N + 1; j <= i; j++ {
				if j >= 0 && j < len(S) && S[j] {
					result[i] = float64(i - j)
					found = true
					break
				}
			}
			if !found {
				result[i] = math.NaN()
			}
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// HHVBARS 求N周期内S最高值到当前周期数
func HHVBARS(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			slice := S[i-N+1 : i+1]
			maxValue := slice[0]
			maxIndex := 0
			for j, v := range slice {
				if v > maxValue {
					maxValue = v
					maxIndex = j
				}
			}
			result[i] = float64(len(slice) - 1 - maxIndex)
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// LLVBARS 求N周期内S最低值到当前周期数
func LLVBARS(S Series, N int) Series {
	result := make(Series, len(S))
	for i := range S {
		if i >= N-1 {
			slice := S[i-N+1 : i+1]
			minValue := slice[0]
			minIndex := 0
			for j, v := range slice {
				if v < minValue {
					minValue = v
					minIndex = j
				}
			}
			result[i] = float64(len(slice) - 1 - minIndex)
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// TOPRANGE 当前最高价是近多少周期内最高价的最大值
func TOPRANGE(S Series) Series {
	result := make(Series, len(S))
	for i := range S {
		if i == 0 {
			result[i] = 0
		} else {
			count := 0
			for j := i - 1; j >= 0; j-- {
				if S[j] < S[i] {
					count++
				} else {
					break
				}
			}
			result[i] = float64(count)
		}
	}
	return result
}

// LOWRANGE 当前最低价是近多少周期内最低价的最小值
func LOWRANGE(S Series) Series {
	result := make(Series, len(S))
	for i := range S {
		if i == 0 {
			result[i] = 0
		} else {
			count := 0
			for j := i - 1; j >= 0; j-- {
				if S[j] > S[i] {
					count++
				} else {
					break
				}
			}
			result[i] = float64(count)
		}
	}
	return result
}

// ------------------ 高级函数版本 ---------------------------------

// SAR 抛物转向指标
func SAR(HIGH, LOW Series, N, S, M int) Series {
	fStep := float64(S) / 100.0
	fMax := float64(M) / 100.0
	af := 0.0
	isLong := HIGH[N-1] > HIGH[N-2]
	bFirst := true
	length := len(HIGH)

	sHHV := REF(HHV(HIGH, N), 1)
	sLLV := REF(LLV(LOW, N), 1)
	sarX := make(Series, length)

	for i := range sarX {
		sarX[i] = math.NaN()
	}

	for i := N; i < length; i++ {
		if bFirst {
			af = fStep
			if isLong {
				sarX[i] = sLLV[i]
			} else {
				sarX[i] = sHHV[i]
			}
			bFirst = false
		} else {
			var ep float64
			if isLong {
				ep = sHHV[i]
			} else {
				ep = sLLV[i]
			}

			if (isLong && HIGH[i] > ep) || (!isLong && LOW[i] < ep) {
				af = math.Min(af+fStep, fMax)
			}

			sarX[i] = sarX[i-1] + af*(ep-sarX[i-1])
		}

		if (isLong && LOW[i] < sarX[i]) || (!isLong && HIGH[i] > sarX[i]) {
			isLong = !isLong
			bFirst = true
		}
	}
	return sarX
}

// TDX_SAR 通达信SAR算法
func TDX_SAR(High, Low Series, iAFStep, iAFLimit int) Series {
	afStep := float64(iAFStep) / 100.0
	afLimit := float64(iAFLimit) / 100.0
	SarX := make(Series, len(High))

	bull := true
	af := afStep
	ep := High[0]
	SarX[0] = Low[0]

	for i := 1; i < len(High); i++ {
		if bull {
			if High[i] > ep {
				ep = High[i]
				af = math.Min(af+afStep, afLimit)
			}
		} else {
			if Low[i] < ep {
				ep = Low[i]
				af = math.Min(af+afStep, afLimit)
			}
		}

		SarX[i] = SarX[i-1] + af*(ep-SarX[i-1])

		if bull {
			SarX[i] = math.Max(SarX[i-1], math.Min(SarX[i], math.Min(Low[i], Low[i-1])))
		} else {
			SarX[i] = math.Min(SarX[i-1], math.Max(SarX[i], math.Max(High[i], High[i-1])))
		}

		if bull {
			if Low[i] < SarX[i] {
				bull = false
				tmpSarX := ep
				ep = Low[i]
				af = afStep
				if High[i-1] == tmpSarX {
					SarX[i] = tmpSarX
				} else {
					SarX[i] = tmpSarX + af*(ep-tmpSarX)
				}
			}
		} else {
			if High[i] > SarX[i] {
				bull = true
				ep = High[i]
				af = afStep
				SarX[i] = math.Min(Low[i], Low[i-1])
			}
		}
	}
	return SarX
}

// QRR 量比
func QRR(VOL Series) Series {
	return DIV(VOL, MA(REF(VOL, 5), 5))
}

// SHO 钱龙短线指标
func SHO(CLOSE, VOL Series, N int) (Series, Series) {
	VAR1 := MA(DIV(SUB(VOL, REF(VOL, 1)), REF(VOL, 1)), 5)
	VAR2 := DIV(MULS(SUB(CLOSE, MA(CLOSE, 24)), 100), MA(CLOSE, 24))
	SHT := MULS(VAR2, 1+VAR1[0])
	SHTMA := MA(SHT, N)
	return SHT, SHTMA
}

// LON 钱龙长线指标
func LON(CLOSE, HIGH, LOW, VOL Series) (Series, Series) {
	LC := REF(CLOSE, 1)
	VID := DIV(SUM(VOL, 2), MULS(SUB(HHV(HIGH, 2), LLV(LOW, 2)), 100))
	RC := MUL(SUB(CLOSE, LC), VID)
	LONG := SUM(RC, 0)
	DIFF := SMA(LONG, 10, 1)
	DEA := SMA(LONG, 20, 1)
	LON := SUB(DIFF, DEA)
	LONMA := MA(LON, 10)
	return LON, LONMA
}
func CopySlice(Slice Series) Series {
	result := make(Series, len(Slice))
	copy(result, Slice)
	return result
}

func DTPRICE(Slice Series, n float64) Series {
	// for i
	w := CopySlice(Slice)
	for i := 0; i < len(w); i++ {
		w[i] = w[i] * (1 - n)
	}
	return w
}
func ZTPRICE(Slice Series, n float64) Series {
	// for i
	w := CopySlice(Slice)
	for i := 0; i < len(w); i++ {
		w[i] = w[i] * (1 + n)
	}
	return w
}

// 全局函数映射表
var functionMap map[string]any

// init 初始化函数映射表
func init() {
	functionMap = map[string]any{
		"DTPRICE": DTPRICE,
		"ZTPRICE": ZTPRICE,
		"IF":      IF,
		// 0级核心工具函数
		"RD":      RD,
		"RET":     RET,
		"ABS":     ABS,
		"LN":      LN,
		"POW":     POW,
		"SQRT":    SQRT,
		"SIN":     SIN,
		"COS":     COS,
		"TAN":     TAN,
		"MAX":     MAX,
		"MIN":     MIN,
		"REF":     REF,
		"DIFF":    DIFF,
		"STD":     STD,
		"SUM":     SUM,
		"MA":      MA,
		"EMA":     EMA,
		"SMA":     SMA,
		"WMA":     WMA,
		"DMA":     DMA,
		"HHV":     HHV,
		"LLV":     LLV,
		"AVEDEV":  AVEDEV,
		"SLOPE":   SLOPE,
		"FORCAST": FORCAST,

		// 1级应用层函数
		"COUNT":      COUNT,
		"EVERY":      EVERY,
		"EXIST":      EXIST,
		"CROSS":      CROSS,
		"LONGCROSS":  LONGCROSS,
		"BARSLAST":   BARSLAST,
		"VALUEWHEN":  VALUEWHEN,
		"BETWEEN":    BETWEEN,
		"FILTER":     FILTER,
		"LAST":       LAST,
		"BARSSINCEN": BARSSINCEN,
		"HHVBARS":    HHVBARS,
		"LLVBARS":    LLVBARS,
		"TOPRANGE":   TOPRANGE,
		"LOWRANGE":   LOWRANGE,

		// 序列运算函数
		"ADD": ADD,
		"SUB": SUB,
		"MUL": MUL,
		"DIV": DIV,

		// 比较函数
		"GreaterThan":        GreaterThan,
		"LessThan":           LessThan,
		"LessThanOrEqual":    LessThanOrEqual,
		"Equal":              Equal,
		"NotEqual":           NotEqual,
		"GreaterThanOrEqual": GreaterThanOrEqual,

		// 2级技术指标函数
		"MACD":  MACD,
		"RSI":   RSI,
		"BOLL":  BOLL,
		"KDJ":   KDJ,
		"WR":    WR,
		"BIAS":  BIAS,
		"PSY":   PSY,
		"CCI":   CCI,
		"ATR":   ATR,
		"BBI":   BBI,
		"DMI":   DMI,
		"TAQ":   TAQ,
		"KTN":   KTN,
		"TRIX":  TRIX,
		"VR":    VR,
		"CR":    CR,
		"EMV":   EMV,
		"DPO":   DPO,
		"BRAR":  BRAR,
		"DFMA":  DFMA,
		"MTM":   MTM,
		"MASS":  MASS,
		"ROC":   ROC,
		"EXPMA": EXPMA,
		"OBV":   OBV,
		"MFI":   MFI,
		"ASI":   ASI,
		"XSII":  XSII,

		// 高级函数版本
		"SAR":     SAR,
		"TDX_SAR": TDX_SAR,
		"QRR":     QRR,
		"SHO":     SHO,
		"LON":     LON,
	}
}

// CallIndicatorByReflection 通过反射直接调用原始函数
func CallIndicatorByReflection(funcName string, args []any) (any, error) {
	// 获取原始函数
	originalFunc := getOriginalFunction(funcName)
	if originalFunc == nil {
		return nil, fmt.Errorf("original function for '%s' not found", funcName)
	}

	// 使用反射调用原始函数
	return callFunctionByReflection(originalFunc, args)
}

// getOriginalFunction 获取原始函数
func getOriginalFunction(funcName string) any {
	if fn, exists := functionMap[funcName]; exists {
		return fn
	}
	return nil
}

// callFunctionByReflection 通过反射调用函数
func callFunctionByReflection(fn any, args []any) (any, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// 检查参数数量
	if fnType.NumIn() != len(args) {
		return nil, fmt.Errorf("argument count mismatch: expected %d, got %d", fnType.NumIn(), len(args))
	}

	// 转换参数类型
	convertedArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		expectedType := fnType.In(i)
		argValue := reflect.ValueOf(arg)

		// 尝试类型转换
		if !argValue.Type().AssignableTo(expectedType) {
			// 尝试自定义类型转换
			converted, err := convertType(argValue, expectedType)
			if err != nil {
				return nil, fmt.Errorf("cannot convert argument %d from %s to %s: %v", i, argValue.Type(), expectedType, err)
			}
			argValue = converted
		}

		convertedArgs[i] = argValue
	}

	// 调用函数
	results := fnValue.Call(convertedArgs)

	// 处理返回值
	if len(results) == 0 {
		return nil, nil
	} else if len(results) == 1 {
		return results[0].Interface(), nil
	} else {
		// 多个返回值，返回切片
		returnValues := make([]any, len(results))
		for i, result := range results {
			returnValues[i] = result.Interface()
		}
		return returnValues, nil
	}
}

// convertType 自定义类型转换，支持 []float64 和 []bool 之间的相互转换
func convertType(src reflect.Value, dstType reflect.Type) (reflect.Value, error) {
	srcType := src.Type()

	// 如果类型可以标准转换，直接使用
	if src.CanConvert(dstType) {
		return src.Convert(dstType), nil
	}

	// 处理 []float64 -> []bool 的转换
	if srcType.Kind() == reflect.Slice && dstType.Kind() == reflect.Slice {
		srcElemType := srcType.Elem()
		dstElemType := dstType.Elem()

		// []float64 -> []bool (包括 Series 类型)
		if srcElemType.Kind() == reflect.Float64 && dstElemType.Kind() == reflect.Bool {
			// 检查是否是 Series 类型
			if src.CanInterface() {
				srcSlice := src.Interface()
				// 尝试转换为 Series
				if series, ok := srcSlice.(Series); ok {
					dstSlice := make([]bool, len(series))
					for i, v := range series {
						dstSlice[i] = v != 0.0
					}
					return reflect.ValueOf(dstSlice), nil
				}
				// 尝试转换为 []float64
				if floatSlice, ok := srcSlice.([]float64); ok {
					dstSlice := make([]bool, len(floatSlice))
					for i, v := range floatSlice {
						dstSlice[i] = v != 0.0
					}
					return reflect.ValueOf(dstSlice), nil
				}
			}
		}

		// []bool -> []float64
		if srcElemType.Kind() == reflect.Bool && dstElemType.Kind() == reflect.Float64 {
			if src.CanInterface() {
				if boolSlice, ok := src.Interface().([]bool); ok {
					dstSlice := make([]float64, len(boolSlice))
					for i, v := range boolSlice {
						if v {
							dstSlice[i] = 1.0
						} else {
							dstSlice[i] = 0.0
						}
					}
					return reflect.ValueOf(dstSlice), nil
				}
			}
		}
	}

	return reflect.Value{}, fmt.Errorf("unsupported type conversion")
}

// 初始化指标函数映射

func GetAllFuncNames() []string {
	// 从 functionMap 中获取所有函数名
	funcNames := make([]string, 0, len(functionMap))
	for funcName := range functionMap {
		funcNames = append(funcNames, funcName)
	}

	// 添加 Series 类型方法（这些不在 functionMap 中）
	seriesMethods := []string{"Len", "At", "Last", "Slice"}
	funcNames = append(funcNames, seriesMethods...)

	return funcNames
}
