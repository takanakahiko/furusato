package furusato_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	furusato "github.com/takanakahiko/furusato/internal"
)

func TestFurusatonozeiLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		input                    furusato.TaxCalculationInput
		wantLimit                int
		wantIncomeTaxDeduction   int
		wantResidentTaxDeduction int
	}{
		{
			name: "年収500万・独身・控除なし",
			input: furusato.TaxCalculationInput{
				SalaryIncome:      5_000_000,
				SocialInsurance:   700_000,
				DeclarationMethod: furusato.NoneDeclaration,
			},
			wantLimit:                62_452,
			wantIncomeTaxDeduction:   6_200,
			wantResidentTaxDeduction: 54_300,
		},
		{
			name: "年収700万・独身・控除なし",
			input: furusato.TaxCalculationInput{
				SalaryIncome:      7_000_000,
				SocialInsurance:   1_000_000,
				DeclarationMethod: furusato.NoneDeclaration,
			},
			wantLimit:                109_943,
			wantIncomeTaxDeduction:   22_100,
			wantResidentTaxDeduction: 85_900,
		},
		{
			name: "年収1000万・独身・各種控除あり",
			input: furusato.TaxCalculationInput{
				SalaryIncome:         10_404_928,
				SocialInsurance:      1_199_515,
				EarthquakeInsurance:  31_302,
				LifeInsuranceGeneral: 100_000,
				HousingLoanDeduction: 280_000,
				DeclarationMethod:    furusato.ElectronicDeclaration,
			},
			wantLimit:                196_723,
			wantIncomeTaxDeduction:   39_800,
			wantResidentTaxDeduction: 154_900,
		},
		{
			name: "年収400万・配偶者控除あり",
			input: furusato.TaxCalculationInput{
				SalaryIncome:      4_000_000,
				SocialInsurance:   600_000,
				SpouseDeduction:   true,
				DeclarationMethod: furusato.NoneDeclaration,
			},
			wantLimit:                33_294,
			wantIncomeTaxDeduction:   1_600,
			wantResidentTaxDeduction: 29_700,
		},
		{
			name: "年収600万・扶養1人",
			input: furusato.TaxCalculationInput{
				SalaryIncome:      6_000_000,
				SocialInsurance:   850_000,
				DependentCount:    1,
				DeclarationMethod: furusato.NoneDeclaration,
			},
			wantLimit:                69_222,
			wantIncomeTaxDeduction:   6_900,
			wantResidentTaxDeduction: 60_300,
		},
		{
			name: "年収2000万・事業所得あり・各種控除フル",
			input: furusato.TaxCalculationInput{
				SalaryIncome:         20_000_000,
				BusinessIncome:       1_000_000,
				SocialInsurance:      2_000_000,
				MedicalExpenses:      300_000,
				EarthquakeInsurance:  50_000,
				LifeInsuranceGeneral: 100_000,
				LifeInsuranceMedical: 60_000,
				LifeInsurancePension: 80_000,
				DeclarationMethod:    furusato.ElectronicDeclaration,
			},
			wantLimit:                559_513,
			wantIncomeTaxDeduction:   188_000,
			wantResidentTaxDeduction: 369_700,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limit := furusato.FurusatonozeiLimit(tt.input)
			incomeTaxDeduction := tt.input.FurusatoDeductionOfIncomeTax(limit)
			residentTaxDeduction := tt.input.FurusatoDeductionOfResidentTax(limit)

			assert.Equal(t, tt.wantLimit, limit, "FurusatonozeiLimit")
			assert.Equal(t, tt.wantIncomeTaxDeduction, incomeTaxDeduction, "FurusatoDeductionOfIncomeTax")
			assert.Equal(t, tt.wantResidentTaxDeduction, residentTaxDeduction, "FurusatoDeductionOfResidentTax")
		})
	}
}
