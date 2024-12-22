package furusato

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// declarationMethod is 申請方法.
type DeclarationMethod int

const (
	// NoneDeclaration is 青色確定申告なし.
	NoneDeclaration DeclarationMethod = iota

	//  ElectronicDeclaration is 電子申告+電子帳簿保存.
	ElectronicDeclaration

	// ElectronicDeclaration is 紙帳簿.
	PaperDeclaration

	// SimpleDeclaration is 簡易帳簿.
	SimpleDeclaration
)

var errInvalidValue = errors.New("invalid value")

func (m *DeclarationMethod) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		return fmt.Errorf("failed to value.Decode: %w", err)
	}

	switch str {
	case "none":
		*m = NoneDeclaration
	case "electronic":
		*m = ElectronicDeclaration
	case "paper":
		*m = PaperDeclaration
	case "simple":
		*m = SimpleDeclaration
	default:
		return fmt.Errorf("invalid declaration method: %w", errInvalidValue)
	}

	return nil
}

// TaxCalculationInput is 入力データをまとめる構造体.
type TaxCalculationInput struct {
	// SalaryIncome is 給与収入(源泉徴収票の"支払金額")
	SalaryIncome int `yaml:"salaryIncome"`
	// MiscellaneousIncome is 雑所得
	MiscellaneousIncome int `yaml:"miscellaneousIncome"`
	// BusinessIncome is 事業所得
	BusinessIncome int `yaml:"businessIncome"`
	// MedicalExpenses is 医療費
	MedicalExpenses int `yaml:"medicalExpenses"`
	// SocialInsurance is 社会保険料控除
	SocialInsurance int `yaml:"socialInsurance"`
	// DependentCount is 扶養親族数(一般の控除対象扶養親族)
	DependentCount int `yaml:"dependentCount"`
	// SpouseDeduction is 配偶者控除の適用有無
	SpouseDeduction bool `yaml:"spouseDeduction"`
	// Method is 申告方法
	DeclarationMethod DeclarationMethod `yaml:"declarationMethod"`

	// furusatoAmount is 内部的に減税効果を算出するためのふるさと納税額
	furusatoAmount int
}

// LoadInput はTaxCalculationInputをyamlから読み込む.
func LoadInput(path string) (TaxCalculationInput, error) {
	// 設定ファイルの読み込み
	file, err := os.Open(path)
	if err != nil {
		return TaxCalculationInput{}, fmt.Errorf("os.Open err: %w", err)
	}
	defer file.Close()

	var input TaxCalculationInput
	if err := yaml.NewDecoder(file).Decode(&input); err != nil {
		return TaxCalculationInput{}, fmt.Errorf("yaml.NewDecoder().Decode err: %w", err)
	}

	return input, nil
}

// PrintInput はTaxCalculationInputをよしなに表示する.
func PrintInput(input TaxCalculationInput) error {
	if out, err := yaml.Marshal(input); err != nil {
		return fmt.Errorf("yaml.Marshal err: %w", err)
	} else {
		fmt.Printf("%s", string(out))
	}

	return nil
}
