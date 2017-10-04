package main

import (
	"fmt"

	"github.com/solvent-io/zps/zps"
)

func main() {
	v1 := &zps.Version{}
	v1.Parse("0.0.1:20160614T141759Z")

	v2 := &zps.Version{}
	v2.Parse("0.1.1:20160613T141759Z")

	v3 := &zps.Version{}
	v3.Parse("0.2.1:20160612T141759Z")

	nachoReq := zps.NewRequirement("fruit", nil).Depends().ANY()
	fruitReq := zps.NewRequirement("bacon", nil).Depends().ANY()

	p1, _ := zps.NewPkg("zpkg://solvent/food/fruit@0.0.1:20160614T141759Z", []*zps.Requirement{fruitReq}, "x86_64", "darwin", "fruit", "fruit")
	p2, _ := zps.NewPkg("zpkg://solvent/food/nacho@0.1.1:20160613T141759Z", []*zps.Requirement{nachoReq}, "x86_64", "darwin", "nacho", "nacho")
	p3, _ := zps.NewPkg("zpkg://solvent/food/bacon@0.2.1:20160612T141759Z", nil, "x86_64", "darwin", "bacon", "bacon")

	p17, _ := zps.NewPkg("zpkg://solvent/food/fruit@0.1.1:20160614T141759Z", []*zps.Requirement{fruitReq}, "x86_64", "darwin", "fruit", "fruit")
	p18, _ := zps.NewPkg("zpkg://solvent/food/nacho@0.1.2:20160613T141759Z", []*zps.Requirement{nachoReq}, "x86_64", "darwin", "nacho", "nacho")
	p19, _ := zps.NewPkg("zpkg://solvent/food/bacon@0.2.1:20160612T141759Z", nil, "x86_64", "darwin", "bacon", "bacon")

	p11, _ := zps.NewPkg("zpkg://solvent/food/fruit@0.2.1:20160612T141759Z", []*zps.Requirement{fruitReq}, "x86_64", "darwin", "fruit", "fruit")
	p12, _ := zps.NewPkg("zpkg://solvent/food/nacho@0.0.1:20160614T141759Z", []*zps.Requirement{nachoReq}, "x86_64", "darwin", "nacho", "nacho")
	p13, _ := zps.NewPkg("zpkg://solvent/food/bacon@0.1.1:20160613T141759Z", nil, "x86_64", "darwin", "bacon", "bacon")

	p14, _ := zps.NewPkg("zpkg://solvent/food/fruit@0.2.1:20160612T141759Z", []*zps.Requirement{fruitReq}, "x86_64", "darwin", "fruit", "fruit")
	p15, _ := zps.NewPkg("zpkg://solvent/food/nacho@0.0.1:20160614T141759Z", []*zps.Requirement{nachoReq}, "x86_64", "darwin", "nacho", "nacho")
	p16, _ := zps.NewPkg("zpkg://solvent/food/bacon@0.1.1:20160613T141759Z", nil, "x86_64", "darwin", "bacon", "bacon")

	p4, _ := zps.NewPkg("zpkg://solvent/animals/pig@0.0.1:20160614T141759Z", nil, "x86_64", "darwin", "pig", "pig")
	p5, _ := zps.NewPkg("zpkg://solvent/animals/cow@0.1.1:20160613T141759Z", nil, "x86_64", "darwin", "cow", "cow")
	p6, _ := zps.NewPkg("zpkg://solvent/animals/duck@0.2.1:20160612T141759Z", nil, "x86_64", "darwin", "duck", "duck")

	p7, _ := zps.NewPkg("zpkg://solvent/people/idiot@0.0.1:20160614T141759Z", nil, "x86_64", "darwin", "idiot", "idiot")
	p8, _ := zps.NewPkg("zpkg://solvent/people/genius@0.1.1:20160613T141759Z", nil, "x86_64", "darwin", "genius", "genius")
	p9, _ := zps.NewPkg("zpkg://solvent/people/zealot@0.2.1:20160612T141759Z", nil, "x86_64", "darwin", "zealot", "zealot")
	p20, _ := zps.NewPkg("zpkg://solvent/food/bacon@0.1.1:20160613T141759Z", nil, "x86_64", "darwin", "bacon", "bacon")
	repo1 := zps.NewRepo("someuri", 0, true, []zps.Solvable{p1, p2, p3, p11, p12, p13})
	repo2 := zps.NewRepo("someuri", 1, true, []zps.Solvable{p17, p18, p19, p14, p15, p16})
	repo3 := zps.NewRepo("someuri", 1, true, []zps.Solvable{p4, p5, p6})

	image := zps.NewRepo("someuri", -1, true, []zps.Solvable{p7, p8, p9, p20})

	pool, err := zps.NewPool(image, repo1, repo2, repo3)
	if err != nil {
		fmt.Println("pool exploded")
	}

	for _, pkg := range pool.Solvables {
		fmt.Println(pkg.Id(), pkg.Priority(), pkg.Location())
	}

	fmt.Println("Pool contains", p5.Name(), pool.Contains(p5))

	dep := zps.NewRequirement("fruit", nil).Depends().ANY()
	candidates := pool.WhatProvides(dep)

	fmt.Println("Packages that provide", dep.Name)
	for _, can := range candidates {
		fmt.Println(can.Id())
	}

	fmt.Println("Request Job")

	request := zps.NewRequest()
	req, err := zps.NewRequirementFromSimpleString("nacho")
	if err != nil {
		fmt.Println(err)
	}

	request.Install(req)

	fmt.Println("Jobs")
	for _, job := range request.Jobs() {
		fmt.Println(job.Op(), job.Requirement().String())
	}

	solver := zps.NewSolver(pool, zps.NewPolicy("updated"))
	solution, err := solver.Solve(request)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("CNF")
	fmt.Println(solver.Cnf())
	fmt.Println("")
	fmt.Println("Sat Solutions")
	for _, sol := range solver.SatSolutions() {
		sol.Print()
		fmt.Println()
	}

	fmt.Println("")
	fmt.Println("Solutions")
	for index, sol := range solver.Solutions() {
		fmt.Println(index, ":")
		for _, sop := range sol.Operations() {
			fmt.Println(sop.Operation, " ", sop.Package.Id(), " ", sop.Package.Priority())
		}
	}

	fmt.Println("")
	fmt.Println("Solution")
	for _, sop := range solution.Operations() {
		fmt.Println(sop.Operation, " ", sop.Package.Id(), " ", sop.Package.Priority())
	}
}
