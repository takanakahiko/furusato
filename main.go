package main

import (
	"fmt"
	"math"
)

const (
	// residentTaxRate = 0.10025           // 住民税率(神奈川県)
	residentTaxRate = 0.1 // 簡素化のため

	specialIncomeTaxRateForReconstruction = 0.021 // 復興特別所得税
	// specialIncomeTaxRateForReconstruction = 0.0 // 簡素化のため

	incomeTaxBasicDeduction   = 480_000 // 所得税の基礎控除
	residentTaxBasicDeduction = 430_000 // 住民税の基礎控除
)

// 申請方法
type declarationMethod string

const (
	ElectronicDeclaration declarationMethod = "電子申告+電子帳簿保存"
	PaperDeclaration      declarationMethod = "紙帳簿"
	SimpleDeclaration     declarationMethod = "簡易帳簿"
)

// 青色申告控除額
func (method declarationMethod) BlueDeduction(businessIncome int) int {
	baseDeduction := 0
	switch method {
	case ElectronicDeclaration:
		baseDeduction = 650_000
	case PaperDeclaration:
		baseDeduction = 550_000
	case SimpleDeclaration:
		baseDeduction = 100_000
	default:
		fmt.Println("不正な申請方法が指定されました。簡易帳簿（10万円控除）を適用します。")
		baseDeduction = 100_000
	}
	if businessIncome < baseDeduction {
		return businessIncome
	}
	return baseDeduction
}

// 入力データをまとめる構造体
type TaxCalculationInput struct {
	SalaryIncome        int               // 給与収入(源泉徴収票の"支払金額")
	MiscellaneousIncome int               // 雑所得
	BusinessIncome      int               // 事業所得
	MedicalExpenses     int               // 医療費
	SocialInsurance     int               // 社会保険料控除
	DependentCount      int               // 扶養親族数(一般の控除対象扶養親族)
	SpouseDeduction     bool              // 配偶者控除の適用有無
	Method              declarationMethod // 申告方法

	furusatoAmount int // 内部的に効果を算出するためのふるさと納税額
}

// 給与所得控除額
func salaryIncomeDeduction(income int) int {
	switch {
	case income <= 1_625_000:
		return 550_000
	case income <= 1_800_000:
		return int(float64(income)*0.4) - 100_000
	case income <= 3_600_000:
		return int(float64(income)*0.3) - 80_000
	case income <= 6_600_000:
		return int(float64(income)*0.2) + 440_000
	case income <= 8_500_000:
		return int(float64(income)*0.1) + 1_100_000
	default:
		return 1_950_000
	}
}

// 総所得(給与所得控除済み)
func (t TaxCalculationInput) TotalIncome() int {
	return t.SalaryIncome - salaryIncomeDeduction(t.SalaryIncome) + t.MiscellaneousIncome + t.BusinessIncome
}

// 医療費控除を計算
func (t TaxCalculationInput) MedicalDeduction() int {
	// 医療費控除
	totalIncome := t.TotalIncome()

	// 医療費控除のしきい値（所得の5%または10万円のいずれか低い方）
	threshold := int(math.Min(float64(totalIncome)*0.05, 100_000))

	// 医療費控除
	medicalDeduction := int(t.MedicalExpenses) - threshold
	if medicalDeduction < 0 {
		return 0
	}
	return medicalDeduction
}

// 課税所得
func (t TaxCalculationInput) TaxableIncome(basicDeduction int) int {
	// 医療費控除
	medicalDeduction := t.MedicalDeduction()

	// 申告特別控除
	blueDeduction := t.Method.BlueDeduction(t.BusinessIncome)

	// 扶養控除
	dependentDeduction := t.DependentCount * 380_000 // TODO 特定扶養親族などの分は別途計算が必要

	// 配偶者控除
	spouseDeduction := 0
	if t.SpouseDeduction {
		spouseDeduction = 380_000
	}

	// 課税所得
	taxableIncome := t.TotalIncome() - medicalDeduction - blueDeduction - t.SocialInsurance - dependentDeduction - spouseDeduction - basicDeduction
	if taxableIncome < 0 {
		return 0
	}
	return taxableIncome
}

// 所得税率と控除額(所得税には復興特別所得税を含まない)
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/2260.htm
func (t TaxCalculationInput) CalculateIncomeTaxRate(basicDeduction int) (rate float64, deduction int) {
	taxableIncome := t.TaxableIncome(basicDeduction)
	switch {
	case taxableIncome <= 1_949_000:
		rate = 0.05
		deduction = 0
	case taxableIncome <= 3_299_999:
		rate = 0.10
		deduction = 97_500
	case taxableIncome <= 6_949_999:
		rate = 0.20
		deduction = 427_500
	case taxableIncome <= 8_999_999:
		rate = 0.23
		deduction = 636_000
	case taxableIncome <= 17_999_999:
		rate = 0.33
		deduction = 1_536_000
	case taxableIncome <= 39_999_999:
		rate = 0.40
		deduction = 2_796_000
	default:
		rate = 0.45
		deduction = 4_796_000
	}
	return rate, deduction
}

// 所得税(復興特別所得税を含む)
func (t TaxCalculationInput) IncomeTax() int {
	// 課税所得
	taxableIncome := t.TaxableIncome(incomeTaxBasicDeduction)

	// ふるさと納税の控除
	// 所得税からの控除をする場合は税から控除するのではなく、課税所得から控除する（後の短数切り捨てに巻き込まれる）
	// https://www.nta.go.jp/publication/pamph/koho/kurashi/html/04_3.htm
	if t.furusatoAmount > 0 {
		taxableIncome = taxableIncome - (t.furusatoAmount - 2000)
	}

	// 課税所得金額を1000円未満の端数切り捨て
	// https://www.nta.go.jp/publication/pamph/koho/kurashi/html/02_1.htm
	// > 1,000円未満切捨て
	taxableIncome = taxableIncome / 1000 * 1000 // 1000円未満の端数切り捨て

	// 特別所得税を含めた所得税率と控除額
	rate, deduction := t.CalculateIncomeTaxRate(incomeTaxBasicDeduction)

	incomeTax := int(float64(taxableIncome)*rate) - deduction
	if incomeTax < 0 {
		return 0
	}

	incomeTax = incomeTax + int(float64(incomeTax)*specialIncomeTaxRateForReconstruction)

	// 定額減税
	// incomeTax = incomeTax - 30000

	// 100円未満の端数切り捨て
	// https://www.nta.go.jp/publication/pamph/koho/kurashi/html/01_1.htm
	// > 100円未満切捨て
	incomeTax = incomeTax / 100 * 100 // 100円未満の端数切り捨て

	return incomeTax
}

// 住民税所得割額(均等割額は含まない)
func (t TaxCalculationInput) ResidentTax(noFurusato bool) int {
	// TODO: 調整控除は面倒で計算していません
	// https://www.city.itabashi.tokyo.jp/tetsuduki/zei/kuminzei/1001751.html
	adjastmentDeduction := 2500 // 調整控除

	// 課税所得
	taxableIncome := t.TaxableIncome(residentTaxBasicDeduction)

	// 住民税の課税標準額
	// 1000円未満の端数切り捨て
	// https://www.city.tokyo-nakano.lg.jp/kurashi/zeikin/jyuminzei-kazei/jyuminzei-keisanrei.html
	// > 課税標準額は、所得合計額から所得控除合計額を差し引いた額（1,000円未満切り捨て）です。
	taxableIncome = taxableIncome / 1000 * 1000

	// 住民税所得割額
	residentTax := int(float64(taxableIncome)*residentTaxRate) - adjastmentDeduction

	// ふるさと納税の控除
	// 住民税からの控除をする場合は所得から控除するのではなく、税から控除する
	// https://www.city.yokohama.lg.jp/kurashi/koseki-zei-hoken/zeikin/y-shizei/kojin-shiminzei-kenminzei/kojin-shiminzei-shosai/zeigakukoujo.html
	if t.furusatoAmount > 0 && !noFurusato {
		incomeTaxRate, _ := t.CalculateIncomeTaxRate(incomeTaxBasicDeduction)
		incomeTaxRateWithForReconstruction := incomeTaxRate * (1 + specialIncomeTaxRateForReconstruction)
		residentTax = residentTax - int(float64(t.furusatoAmount-2000)*residentTaxRate)

		// TODO: 限度額を超えた場合のロジックを書いたほうがいい
		residentTax = residentTax - int(float64(t.furusatoAmount-2000)*(1.0-residentTaxRate-incomeTaxRateWithForReconstruction))
	}

	residentTax = residentTax / 100 * 100 // 100円未満の端数切り捨て

	return residentTax
}

func (t TaxCalculationInput) FurusatoDeductionOfIncomeTax(furusatoAmount int) int {
	t2 := t
	t2.furusatoAmount = furusatoAmount
	return t.IncomeTax() - t2.IncomeTax()
}

func (t TaxCalculationInput) FurusatoDeductionOfResidentTax(furusatoAmount int) int {
	t2 := t
	t2.furusatoAmount = furusatoAmount
	return t.ResidentTax(true) - t2.ResidentTax(false)
}

// ふるさと納税の控除上限額
func (t TaxCalculationInput) FurusatoNozeiLimit() int {
	// 所得税率と控除額
	incomeTaxRate, _ := t.CalculateIncomeTaxRate(incomeTaxBasicDeduction)
	incomeTaxRateWithForReconstruction := incomeTaxRate * (1 + specialIncomeTaxRateForReconstruction)
	fmt.Printf("所得税率: %f\n", incomeTaxRate)

	// 所得税
	incomeTax := t.IncomeTax() // 所得税
	fmt.Printf("所得税にかかる課税所得: %d\n", t.TaxableIncome(incomeTaxBasicDeduction))
	fmt.Printf("所得税: %d\n", incomeTax)

	// 住民税
	residentTax := t.ResidentTax(true) // 住民税所得割額
	fmt.Printf("住民税にかかる課税所得: %d\n", t.TaxableIncome(residentTaxBasicDeduction))
	fmt.Printf("住民税所得割額: %d円\n\n", residentTax)

	// ふるさと納税の控除上限額
	// 住民税特例分からの控除限度額＝個人住民税所得割額の20%
	// 住民税からの控除（特例分） = （ふるさと納税額 - 2,000円）×（100％ - 10％（基本分） - 所得税の税率）
	// 個人住民税所得割額 * 20％ = (X-2,000円) * (90％-所得税の税率)
	// X = 個人住民税所得割額 * 20％ /（90％-所得税の税率）+ 2,000円
	return int(float64(residentTax)*0.2/(1-residentTaxRate-incomeTaxRateWithForReconstruction)) + 2000
}

func main() {
	// 入力データ
	input := TaxCalculationInput{
		SalaryIncome:        5_000_000,
		MiscellaneousIncome: 100_000,
		BusinessIncome:      500_000,
		MedicalExpenses:     150_000,
		SocialInsurance:     600_000,
		DependentCount:      1,
		SpouseDeduction:     true,
		Method:              ElectronicDeclaration,
	}

	// 所得税
	incomeTax := input.IncomeTax()
	fmt.Printf("所得税にかかる課税所得: %d\n", input.TaxableIncome(incomeTaxBasicDeduction))
	fmt.Printf("所得税: %d\n", incomeTax)

	// 住民税
	residentTax := input.ResidentTax(true)
	fmt.Printf("住民税にかかる課税所得: %d\n", input.TaxableIncome(residentTaxBasicDeduction))
	fmt.Printf("住民税所得割額: %d円\n\n", residentTax)

	// ふるさと納税の控除上限額
	limit := input.FurusatoNozeiLimit()
	fmt.Printf("ふるさと納税で使える上限額は: %d円です\n\n", limit)

	// ふるさと納税の控除額
	limit = 63536
	fmt.Printf("所得税からの控除: %d円\n", input.FurusatoDeductionOfIncomeTax(limit))
	fmt.Printf("住民税からの控除: %d円\n", input.FurusatoDeductionOfResidentTax(limit))
}
