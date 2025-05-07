// MIT License
//
// (C) Copyright [2018-2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sharedtest

import (
	base "github.com/Cray-HPE/hms-base/v2"
)

func CompareComp(comp1 *base.Component, comp2 *base.Component) bool {
	return CompareComps([]*base.Component{comp1}, []*base.Component{comp2})
}

func CompareComps(comps1 []*base.Component, comps2 []*base.Component) bool {
	if len(comps1) != len(comps2) {
		return false
	}
	// Treat nil entries specially.  There should be the same number,
	// and the remaining non-nil entries should be checked against
	// each other.
	nils := 0
	foundNils := 0
	for _, c1 := range comps1 {
		if c1 == nil {
			nils += 1
		}
	}
	for _, c2 := range comps2 {
		if c2 == nil {
			foundNils += 1
		}
	}
	if nils != foundNils {
		return false
	}
	for _, c1 := range comps1 {
		if c1 == nil {
			continue
		}
		found := false
		for _, c2 := range comps2 {
			if c2 == nil {
				continue
			}
			if c1.ID != c2.ID {
				continue
			}
			if c1.Type != c2.Type {
				continue
			}
			if c1.State != c2.State {
				continue
			}
			if c1.Flag != c2.Flag {
				continue
			}
			if c1.Enabled != nil {
				if c2.Enabled != nil {
					if *c1.Enabled != *c2.Enabled {
						continue
					}
					// Matches
				} else {
					continue
				}
			} else if c2.Enabled != nil {
				continue
			}
			if c1.SwStatus != c2.SwStatus {
				continue
			}
			if c1.Role != c2.Role {
				continue
			}
			if c1.SubRole != c2.SubRole {
				continue
			}
			if c1.NID != c2.NID {
				continue
			}
			if c1.Subtype != c2.Subtype {
				continue
			}
			if c1.NetType != c2.NetType {
				continue
			}
			if c1.Arch != c2.Arch {
				continue
			}
			if c1.Class != c2.Class {
				continue
			}
			found = true
			break
		}
		if !found {
			return false
		}
	}
	return true
}
