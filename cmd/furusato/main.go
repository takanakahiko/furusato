package main

import (
	"flag"
	"fmt"
	"log"

	furusato "github.com/takanakahiko/furusato/internal"
)

func main() {
	// コマンドライン引数の解析
	inputPath := flag.String("input", "furusato.yml", "yaml input file")
	flag.Parse()

	// 入力データの読み込み
	input, err := furusato.LoadInput(*inputPath)
	if err != nil {
		log.Fatal("LoadInput err", err)
	}

	// 入力データの表示
	if err := furusato.PrintInput(input); err != nil {
		log.Fatal("json.MarshalIndent err", err)
	}

	fmt.Println() // 改行

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
