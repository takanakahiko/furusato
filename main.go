package main

import (
	"fmt"
	"math"
	"slices"
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
	SalaryIncome        int               // 給与所得
	MiscellaneousIncome int               // 雑所得
	BusinessIncome      int               // 事業所得
	MedicalExpenses     int               // 医療費
	SocialInsurance     int               // 社会保険料控除
	DependentCount      int               // 扶養親族数
	SpouseDeduction     bool              // 配偶者控除の適用有無
	Method              declarationMethod // 青色申告申請方法
}

// 医療費控除
func (t TaxCalculationInput) TotalIncome() int {
	return t.SalaryIncome + t.MiscellaneousIncome + t.BusinessIncome
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
func (t TaxCalculationInput) TaxableIncome() int {
	// 医療費控除
	medicalDeduction := t.MedicalDeduction()

	// 申告特別控除
	blueDeduction := t.Method.BlueDeduction(t.BusinessIncome)

	// 扶養控除
	dependentDeduction := t.DependentCount * 380_000

	// 配偶者控除
	spouseDeduction := 0
	if t.SpouseDeduction {
		spouseDeduction = 380_000
	}

	// 基礎控除
	basicDeduction := 480_000

	// 課税所得
	taxableIncome := t.TotalIncome() - medicalDeduction - blueDeduction - t.SocialInsurance - dependentDeduction - spouseDeduction - basicDeduction
	if taxableIncome < 0 {
		return 0
	}
	return taxableIncome
}

// 所得税率と控除額
// https://www.nta.go.jp/taxes/shiraberu/taxanswer/shotoku/2260.htm
func (t TaxCalculationInput) CalculateIncomeTaxRate() (rate float64, deduction int) {
	taxableIncome := t.TaxableIncome()
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

// ふるさと納税の控除上限額
func (t TaxCalculationInput) FurusatoNozeiLimit() int {
	// 課税所得
	taxableIncome := t.TaxableIncome()
	fmt.Printf("taxableIncome: %d円\n", taxableIncome)

	// 所得税率と控除額
	incomeTaxRate, incomeTaxDeduction := t.CalculateIncomeTaxRate()

	// 所得税
	incomeTax := int(float64(taxableIncome)*incomeTaxRate) - incomeTaxDeduction
	if incomeTax < 0 {
		incomeTax = 0
	}

	// 所得税からの控除限度額＝総所得の40％
	// 所得税からの控除 = （ふるさと納税額－2,000円）×「所得税の税率」
	// 総所得 * 40％ = (X-2,000円) * 所得税の税率
	// X = 総所得 * 40% / 所得税の税率 + 2000円
	limitByIncomeTax := int(float64(t.TotalIncome())*0.4/incomeTaxRate) + 2000
	fmt.Printf("limitByIncomeTax: %d円\n", limitByIncomeTax)

	// 住民税基本分からの控除限度額＝総所得の30％
	// 住民税からの控除（基本分） = （ふるさと納税額－2,000円）×10％
	// 総所得 * 30％ = (X-2,000円) * 10％
	// X = 総所得 * 30％以下 / 10％ + 2000円
	limitByResidentTax := int(float64(t.TotalIncome())*0.3/0.1) + 2000
	fmt.Printf("limitByResidentTax: %d円\n", limitByResidentTax)

	// 住民税特例分からの控除限度額＝個人住民税所得割額の20%
	// 住民税からの控除（特例分） = （ふるさと納税額 - 2,000円）×（100％ - 10％（基本分） - 所得税の税率）
	// 住民税所得割額 * 20％ = (X-2,000円) * (90％-所得税の税率)
	// X = 個人住民税所得割額 * 20％ /（90％-所得税の税率）+ 2,000円
	residentTaxRate := 0.10025                                   // 住民税率(神奈川県)
	residentTax := int(float64(taxableIncome) * residentTaxRate) // 住民税所得割額
	fmt.Printf("住民税所得割額: %d円\n", residentTax)
	limitByResidentTaxSpecial := int(float64(residentTax)*0.2/((0.9-incomeTaxRate)*1.021)) + 2000
	fmt.Printf("limitByResidentTaxSpecial: %d円\n\n", limitByResidentTaxSpecial)

	// デバッグ用
	limit := slices.Min([]int{limitByIncomeTax, limitByResidentTax, limitByResidentTaxSpecial})
	fmt.Printf("incomeTaxRate: %f\n", incomeTaxRate)
	fmt.Printf("所得税からの控除: %d\n", int(float64(limit-2000)*incomeTaxRate))
	fmt.Printf("住民税基本分からの控除: %d\n", int(float64(limit-2000)*0.1))
	fmt.Printf("住民税特例分からの控除: %d\n\n", int(float64(limit-2000)*(0.9-incomeTaxRate*1.021)))

	// ふるさと納税の控除上限額
	return limit
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

	limit := input.FurusatoNozeiLimit()
	fmt.Printf("ふるさと納税で使える上限額は: %d円です\n", limit)
}
