// Code generated by "stringer -type=TypeKind -output=type-kind_string.go"; DO NOT EDIT.

package appdef

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeKind_null-0]
	_ = x[TypeKind_GDoc-1]
	_ = x[TypeKind_CDoc-2]
	_ = x[TypeKind_ODoc-3]
	_ = x[TypeKind_WDoc-4]
	_ = x[TypeKind_GRecord-5]
	_ = x[TypeKind_CRecord-6]
	_ = x[TypeKind_ORecord-7]
	_ = x[TypeKind_WRecord-8]
	_ = x[TypeKind_ViewRecord-9]
	_ = x[TypeKind_Object-10]
	_ = x[TypeKind_Element-11]
	_ = x[TypeKind_Query-12]
	_ = x[TypeKind_Command-13]
	_ = x[TypeKind_Workspace-14]
	_ = x[TypeKind_FakeLast-15]
}

const _TypeKind_name = "TypeKind_nullTypeKind_GDocTypeKind_CDocTypeKind_ODocTypeKind_WDocTypeKind_GRecordTypeKind_CRecordTypeKind_ORecordTypeKind_WRecordTypeKind_ViewRecordTypeKind_ObjectTypeKind_ElementTypeKind_QueryTypeKind_CommandTypeKind_WorkspaceTypeKind_FakeLast"

var _TypeKind_index = [...]uint8{0, 13, 26, 39, 52, 65, 81, 97, 113, 129, 148, 163, 179, 193, 209, 227, 244}

func (i TypeKind) String() string {
	if i >= TypeKind(len(_TypeKind_index)-1) {
		return "TypeKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TypeKind_name[_TypeKind_index[i]:_TypeKind_index[i+1]]
}