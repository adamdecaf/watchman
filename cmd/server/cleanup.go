// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"strings"
)

// original list: inc, incorporated, llc, llp, co, ltd, limited, sa de cv, corporation, corp, ltda,
//                open joint stock company, pty ltd, public limited company, ag, cjsc, plc, as, aps,
//                oy, sa, gmbh, se, pvt ltd, sp zoo, ooo, sl, pjsc, jsc, bv, pt, tbk

var (
	companySuffixReplacer = strings.NewReplacer(
		" CO.", "",
		" INC.", "",
		" GMBH", "",
		" LLC", "", " LLP", "",
		" LTD.", "", ", LTD", "", " LTDA.", "",
	)
)

// 24428,"SAI ADVISORS INC.",-0- ,"VENEZUELA",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Company Number 68-0678326 (United States); Linked To: SARRIA DIAZ, Rafael Alfredo."
// 3748,"COBALT REFINERY CO. INC.",-0- ,"CUBA",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0-

// 6953,"AL BARAKA EXCHANGE LLC",-0- ,"SDGT",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0-
// 11589,"RUNNING BROOK, LLC (USA)",-0- ,"SDNTK",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"US FEIN 030510902 (United States); Business Registration Document # L00000010931 (United States)."

// 20259,"YAKIMA OIL TRADING, LLP",-0- ,"SDNTK",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Commercial Registry Number OC390985 (United Kingdom)."

// 21553,"MKS INTERNATIONAL CO. LTD.",-0- ,"NPWMD] [IFSR",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Additional Sanctions Information - Subject to Secondary Sanctions."
// 22246,"SHANGHAI NORTH TRANSWAY INTERNATIONAL TRADING CO.",-0- ,"NPWMD] [IFSR",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Additional Sanctions Information - Subject to Secondary Sanctions."

// 22603,"DANDONG ZHICHENG METALLIC MATERIAL CO., LTD.",-0- ,"DPRK3",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Secondary sanctions risk: North Korea Sanctions Regulations, sections 510.201 and 510.210."
// 8310,"ADVANCED ELECTRONICS DEVELOPMENT, LTD",-0- ,"IRAQ2",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0-
// 8340,"AMD CO. LTD AGENCY",-0- ,"IRAQ2",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0-
// 8397,"REYNOLDS AND WILSON, LTD.",-0- ,"IRAQ2",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0-

// 8732,"AEROCOMERCIAL ALAS DE COLOMBIA LTDA.",-0- ,"SDNT",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"NIT # 800049071-7 (Colombia)."
// 8877,"DIMABE LTDA.",-0- ,"SDNT",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"NIT # 800107988-4 (Colombia)."

// 11613,"ASCOTEC STEEL TRADING GMBH",-0- ,"IRAN",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Additional Sanctions Information - Subject to Secondary Sanctions; Registration ID HRB 48319 (Germany); all offices worldwide."
// 2110,"TROPIC TOURS GMBH",-0- ,"CUBA",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0-

// Controls
// 16006,"TADBIR ECONOMIC DEVELOPMENT GROUP",-0- ,"IRAN",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Additional Sanctions Information - Subject to Secondary Sanctions."
// 16128,"DI LAURO, Marco","individual","TCO",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"DOB 16 Jun 1980; POB Naples, Italy."
// 16136,"PETRO ROYAL FZE",-0- ,"IRAN",-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,-0- ,"Additional Sanctions Information - Subject to Secondary Sanctions."

func removeCompanySuffixes(in string) string {
	return companySuffixReplacer.Replace(in)
}

// require min match length in watchman? (or downweight initials, suffixes: Dr., Sr.)
// rosette: expands nicknames (Dave -> David)
//  - weights by word length, initials 10-15%, last name 49%, first name mid30%
//    - might have just been relative to query length

// Input Name       Max Score     Matched Name            Score   Matched Alt Name   Score    Matched Address   Score
// Juan Cruz        80.6%         MOSQUERA, Juan Carlos   80.6%   Other              20.4%    John Smith        17.3%

// TODO(Adam): work on identifiers to remove in company names/titles
