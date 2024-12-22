package main

import (
	"fmt"

	furusato "github.com/takanakahiko/furusato/internal"
)

func main() {
	// 入力データ(ダミーデータ)
	input := furusato.TaxCalculationInput{
		SalaryIncome:        5_000_000,
		MiscellaneousIncome: 100_000,
		BusinessIncome:      500_000,
		MedicalExpenses:     150_000,
		SocialInsurance:     600_000,
		DependentCount:      1,
		SpouseDeduction:     true,
		Method:              furusato.ElectronicDeclaration,
	}

	// 所得税
	fmt.Printf("所得税にかかる課税所得: %d円\n", furusato.TaxableIncomeForIncomeTax(input))
	fmt.Printf("所得税: %d円\n\n", furusato.IncomeTax(input))

	// 住民税
	fmt.Printf("住民税にかかる課税所得: %d円\n", furusato.TaxableIncomeForResindentTax(input))
	fmt.Printf("住民税所得割額: %d円\n\n", furusato.ResidentTax(input, true))

	// ふるさと納税の控除上限額
	limit := furusato.FurusatoNozeiLimit(input)
	fmt.Printf("ふるさと納税で使える上限額は: %d円\n\n", limit)

	// ふるさと納税の控除額
	fmt.Printf("所得税からの控除: %d円\n", input.FurusatoDeductionOfIncomeTax(limit))
	fmt.Printf("住民税からの控除: %d円\n", input.FurusatoDeductionOfResidentTax(limit))
}
