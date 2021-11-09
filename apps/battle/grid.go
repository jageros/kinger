package main

import "kinger/gopuppy/attribute"

var _ iTarget = &deskGrid{}

type deskGrid struct {
	baseTarget
}

func newGrid(objID, gridID int, situation *battleSituation) *deskGrid {
	dg := &deskGrid{}
	dg.objID = objID
	dg.gridID = gridID
	dg.situation = situation
	dg.targetType = stEmptyGrid
	return dg
}

func (dg *deskGrid) copy(situation *battleSituation) iTarget {
	c := *dg
	cpy := &c
	cpy.situation = situation
	cpy.effects = []*mcMovieEffect{}
	return cpy
}

func (dg *deskGrid) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("attrType", attrGrid)
	attr.SetMapAttr("baseTarget", dg.baseTarget.packAttr())
	return attr
}

func (dg *deskGrid) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	(&dg.baseTarget).restoredFromAttr(attr.GetMapAttr("baseTarget"), situation)
}

func (dg *deskGrid) getCopyTarget() iTarget {
	return dg
}
