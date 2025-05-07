// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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

package sm

import (
	"reflect"
	base "github.com/Cray-HPE/hms-base/v2"
	"testing"
)

//
// Test format checking for database keys and query parameters.
//

func TestVerifyGroupField(t *testing.T) {
	tests := []struct {
		in          string
		expectedOut error
	}{{
		in:          "foo_bar",
		expectedOut: nil,
	}, {
		in:          "f00-b.r:",
		expectedOut: nil,
	}, {
		in:          "foobar!",
		expectedOut: ErrGroupBadField,
	}, {
		in:          "FooBar",
		expectedOut: ErrGroupBadField,
	}}
	for i, test := range tests {
		out := VerifyGroupField(test.in)
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedOut, out)
		}
	}
}

//
// Test members
//

func TestNewMembers(t *testing.T) {
	tests := []struct {
		expectedOut *Members
	}{{
		expectedOut: &Members{IDs: []string{}},
	}}
	for i, test := range tests {
		out := NewMembers()
		if !reflect.DeepEqual(test.expectedOut, out) {
			t.Errorf("Test %v Failed: Expected members struct '%v'; Received members struct '%v'", i, test.expectedOut, out)
		}
	}
}

func TestVerifyMembers(t *testing.T) {
	tests := []struct {
		in          *Members
		expectedOut error
	}{{
		in: &Members{
			IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
		},
		expectedOut: nil,
	}, {
		in: &Members{
			IDs:      []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			verified: true,
		},
		expectedOut: nil,
	}, {
		in:          &Members{IDs: []string{"x0c0s1b0n0", "foo"}},
		expectedOut: base.ErrHMSTypeInvalid,
	}}
	for i, test := range tests {
		out := test.in.Verify()
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedOut, out)
		}
	}
}

func TestNormalizeMembers(t *testing.T) {
	tests := []struct {
		in          *Members
		expectedOut *Members
	}{{
		in: &Members{
			IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
		},
		expectedOut: &Members{
			IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			normalized: true,
		},
	}, {
		in: &Members{
			IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
		},
		expectedOut: &Members{
			IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			normalized: true,
		},
	}, {
		in: &Members{
			IDs:        []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			normalized: true,
		},
		expectedOut: &Members{
			IDs:        []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			normalized: true,
		},
	}}
	for i, test := range tests {
		test.in.Normalize()
		if !reflect.DeepEqual(test.expectedOut, test.in) {
			t.Errorf("Test %v Failed: Expected members struct '%v'; Received members struct '%v'", i, test.expectedOut, test.in)
		}
	}
}

//
// Test groups
//

func TestNewGroup(t *testing.T) {
	tests := []struct {
		labelIn      string
		descIn       string
		exclGrpIn    string
		tagsIn       []string
		member_idsIn []string
		expectedOut  *Group
		expectedErr  error
	}{{
		labelIn:      "my_group",
		descIn:       "This is my group",
		exclGrpIn:    "my_system",
		tagsIn:       []string{"foo", "bar"},
		member_idsIn: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
		expectedOut: &Group{
			Label:          "my_group",
			Description:    "This is my group",
			ExclusiveGroup: "my_system",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
				verified:   true,
				normalized: true,
			},
			verified:   true,
			normalized: true,
		},
		expectedErr: nil,
	}, {
		labelIn:      "MyGroup",
		descIn:       "This is my group",
		exclGrpIn:    "MySystem",
		tagsIn:       []string{"foo", "bar"},
		member_idsIn: []string{"x0c0s01b0n0", "x0c0s02b0n0"},
		expectedOut: &Group{
			Label:          "mygroup",
			Description:    "This is my group",
			ExclusiveGroup: "mysystem",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
				verified:   true,
				normalized: true,
			},
			verified:   true,
			normalized: true,
		},
		expectedErr: nil,
	}, {
		labelIn:      "My Group",
		descIn:       "This is my group",
		exclGrpIn:    "my_system",
		tagsIn:       []string{"foo", "bar"},
		member_idsIn: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
		expectedOut:  &Group{},
		expectedErr:  ErrGroupBadField,
	}}
	for i, test := range tests {
		out, err := NewGroup(test.labelIn, test.descIn, test.exclGrpIn, test.tagsIn, test.member_idsIn)
		if err != test.expectedErr {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedErr, err)
		} else if test.expectedErr == nil {
			if !reflect.DeepEqual(test.expectedOut, out) {
				t.Errorf("Test %v Failed: Expected group struct '%v'; Received group struct '%v'", i, test.expectedOut, out)
			}
		}
	}
}

func TestVerifyGroup(t *testing.T) {
	tests := []struct {
		in          *Group
		expectedOut error
	}{{
		in: &Group{
			Label:          "my_group",
			Description:    "This is my group",
			ExclusiveGroup: "my_system",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: nil,
	}, {
		in: &Group{
			Label:          "MyGroup",
			Description:    "This is my group",
			ExclusiveGroup: "my_system",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
			verified: true,
		},
		expectedOut: nil,
	}, {
		in: &Group{
			Label:          "MyGroup",
			Description:    "This is my group",
			ExclusiveGroup: "my_system",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: ErrGroupBadField,
	}, {
		in: &Group{
			Label:          "my_group",
			Description:    "This is my group",
			ExclusiveGroup: "MySystem",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: ErrGroupBadField,
	}, {
		in: &Group{
			Label:          "my_group",
			Description:    "This is my group",
			ExclusiveGroup: "",
			Tags:           []string{"Foo"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: ErrGroupBadField,
	}, {
		in: &Group{
			Label:          "my_group",
			Description:    "This is my group",
			ExclusiveGroup: "",
			Tags:           []string{},
			Members: Members{
				IDs: []string{"foo", "bar"},
			},
		},
		expectedOut: base.ErrHMSTypeInvalid,
	}}
	for i, test := range tests {
		out := test.in.Verify()
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedOut, out)
		}
	}
}

func TestNormalizeGroup(t *testing.T) {
	tests := []struct {
		in          *Group
		expectedOut *Group
	}{{
		in: &Group{
			Label:          "My_Group",
			Description:    "This is my group",
			ExclusiveGroup: "My_System",
			Tags:           []string{"Foo", "Bar"},
			Members: Members{
				IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: &Group{
			Label:          "my_group",
			Description:    "This is my group",
			ExclusiveGroup: "my_system",
			Tags:           []string{"foo", "bar"},
			Members: Members{
				IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
				normalized: true,
			},
			normalized: true,
		},
	}, {
		in: &Group{
			Label:          "My_Group",
			Description:    "This is my group",
			ExclusiveGroup: "My_System",
			Tags:           []string{"Foo", "Bar"},
			Members: Members{
				IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			},
			normalized: true,
		},
		expectedOut: &Group{
			Label:          "My_Group",
			Description:    "This is my group",
			ExclusiveGroup: "My_System",
			Tags:           []string{"Foo", "Bar"},
			Members: Members{
				IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			},
			normalized: true,
		},
	}}
	for i, test := range tests {
		test.in.Normalize()
		if !reflect.DeepEqual(test.expectedOut, test.in) {
			t.Errorf("Test %v Failed: Expected group struct '%v'; Received group struct '%v'", i, test.expectedOut, test.in)
		}
	}
}

func TestVerifyGroupPatch(t *testing.T) {
	tests := []struct {
		in          *GroupPatch
		expectedOut error
	}{{
		in: &GroupPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedOut: nil,
	}, {
		in: &GroupPatch{
			Tags: &[]string{"Foo", "bar"},
		},
		expectedOut: ErrGroupBadField,
	}}
	for i, test := range tests {
		out := test.in.Verify()
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedOut, out)
		}
	}
}

func TestNormalizeGroupPatch(t *testing.T) {
	tests := []struct {
		in          *GroupPatch
		expectedOut *GroupPatch
	}{{
		in: &GroupPatch{
			Tags: &[]string{"Foo", "Bar"},
		},
		expectedOut: &GroupPatch{
			Tags: &[]string{"foo", "bar"},
		},
	}}
	for i, test := range tests {
		test.in.Normalize()
		if !reflect.DeepEqual(test.expectedOut, test.in) {
			t.Errorf("Test %v Failed: Expected group patch struct '%v'; Received group patch struct '%v'", i, test.expectedOut, test.in)
		}
	}
}

//
// Test partitions
//

func TestNewPartition(t *testing.T) {
	tests := []struct {
		nameIn       string
		descIn       string
		tagsIn       []string
		member_idsIn []string
		expectedOut  *Partition
		expectedErr  error
	}{{
		nameIn:       "p1",
		descIn:       "This is my partition",
		tagsIn:       []string{"foo", "bar"},
		member_idsIn: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
		expectedOut: &Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members: Members{
				IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
				verified:   true,
				normalized: true,
			},
			verified:   true,
			normalized: true,
		},
		expectedErr: nil,
	}, {
		nameIn:       "P1.4",
		descIn:       "This is my partition",
		tagsIn:       []string{"foo", "bar"},
		member_idsIn: []string{"x0c0s01b0n0", "x00c0s02b0n0"},
		expectedOut: &Partition{
			Name:        "p1.4",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members: Members{
				IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
				verified:   true,
				normalized: true,
			},
			verified:   true,
			normalized: true,
		},
		expectedErr: nil,
	}, {
		nameIn:       "p1",
		descIn:       "This is my partition",
		tagsIn:       []string{"foo", "bar!"},
		member_idsIn: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
		expectedOut:  &Partition{},
		expectedErr:  ErrGroupBadField,
	}}
	for i, test := range tests {
		out, err := NewPartition(test.nameIn, test.descIn, test.tagsIn, test.member_idsIn)
		if err != test.expectedErr {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedErr, err)
		} else if test.expectedErr == nil {
			if !reflect.DeepEqual(test.expectedOut, out) {
				t.Errorf("Test %v Failed: Expected partition struct '%v'; Received partition struct '%v'", i, test.expectedOut, out)
			}
		}
	}
}

func TestVerifyPartition(t *testing.T) {
	tests := []struct {
		in          *Partition
		expectedOut error
	}{{
		in: &Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: nil,
	}, {
		in: &Partition{
			Name:        "p1.4",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
			verified: true,
		},
		expectedOut: nil,
	}, {
		in: &Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"Foo"},
			Members: Members{
				IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: ErrGroupBadField,
	}, {
		in: &Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{},
			Members: Members{
				IDs: []string{"foo", "bar"},
			},
		},
		expectedOut: base.ErrHMSTypeInvalid,
	}}
	for i, test := range tests {
		out := test.in.Verify()
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedOut, out)
		}
	}
}

func TestNormalizePartition(t *testing.T) {
	tests := []struct {
		in          *Partition
		expectedOut *Partition
	}{{
		in: &Partition{
			Name:        "My_Part",
			Description: "This is my partition",
			Tags:        []string{"Foo", "Bar"},
			Members: Members{
				IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			},
		},
		expectedOut: &Partition{
			Name:        "my_part",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members: Members{
				IDs:        []string{"x0c0s1b0n0", "x0c0s2b0n0"},
				normalized: true,
			},
			normalized: true,
		},
	}, {
		in: &Partition{
			Name:        "My_Part",
			Description: "This is my partition",
			Tags:        []string{"Foo", "Bar"},
			Members: Members{
				IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			},
			normalized: true,
		},
		expectedOut: &Partition{
			Name:        "My_Part",
			Description: "This is my partition",
			Tags:        []string{"Foo", "Bar"},
			Members: Members{
				IDs: []string{"x0C0s1B0n0", "x0c0s2b0n0"},
			},
			normalized: true,
		},
	}}
	for i, test := range tests {
		test.in.Normalize()
		if !reflect.DeepEqual(test.expectedOut, test.in) {
			t.Errorf("Test %v Failed: Expected partition struct '%v'; Received partition struct '%v'", i, test.expectedOut, test.in)
		}
	}
}

func TestVerifyPartitionPatch(t *testing.T) {
	tests := []struct {
		in          *PartitionPatch
		expectedOut error
	}{{
		in: &PartitionPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedOut: nil,
	}, {
		in: &PartitionPatch{
			Tags: &[]string{"Foo", "bar"},
		},
		expectedOut: ErrGroupBadField,
	}}
	for i, test := range tests {
		out := test.in.Verify()
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected error '%v'; Received error '%v'", i, test.expectedOut, out)
		}
	}
}

func TestNormalizePartitionPatch(t *testing.T) {
	tests := []struct {
		in          *PartitionPatch
		expectedOut *PartitionPatch
	}{{
		in: &PartitionPatch{
			Tags: &[]string{"Foo", "Bar"},
		},
		expectedOut: &PartitionPatch{
			Tags: &[]string{"foo", "bar"},
		},
	}}
	for i, test := range tests {
		test.in.Normalize()
		if !reflect.DeepEqual(test.expectedOut, test.in) {
			t.Errorf("Test %v Failed: Expected partition patch struct '%v'; Received partition patch struct '%v'", i, test.expectedOut, test.in)
		}
	}
}
