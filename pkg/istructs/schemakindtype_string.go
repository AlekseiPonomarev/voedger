// Code generated by "stringer -type=SchemaKindType"; DO NOT EDIT.

package istructs

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SchemaKind_null-0]
	_ = x[SchemaKind_GDoc-1]
	_ = x[SchemaKind_CDoc-2]
	_ = x[SchemaKind_ODoc-3]
	_ = x[SchemaKind_WDoc-4]
	_ = x[SchemaKind_GRecord-5]
	_ = x[SchemaKind_CRecord-6]
	_ = x[SchemaKind_ORecord-7]
	_ = x[SchemaKind_WRecord-8]
	_ = x[SchemaKind_ViewRecord-9]
	_ = x[SchemaKind_ViewRecord_PartitionKey-10]
	_ = x[SchemaKind_ViewRecord_ClusteringColumns-11]
	_ = x[SchemaKind_ViewRecord_Value-12]
	_ = x[SchemaKind_Object-13]
	_ = x[SchemaKind_Element-14]
	_ = x[SchemaKind_QueryFunction-15]
	_ = x[SchemaKind_CommandFunction-16]
	_ = x[SchemaKind_FakeLast-17]
}

const _SchemaKindType_name = "SchemaKind_nullSchemaKind_GDocSchemaKind_CDocSchemaKind_ODocSchemaKind_WDocSchemaKind_GRecordSchemaKind_CRecordSchemaKind_ORecordSchemaKind_WRecordSchemaKind_ViewRecordSchemaKind_ViewRecord_PartitionKeySchemaKind_ViewRecord_ClusteringColumnsSchemaKind_ViewRecord_ValueSchemaKind_ObjectSchemaKind_ElementSchemaKind_QueryFunctionSchemaKind_CommandFunctionSchemaKind_FakeLast"

var _SchemaKindType_index = [...]uint16{0, 15, 30, 45, 60, 75, 93, 111, 129, 147, 168, 202, 241, 268, 285, 303, 327, 353, 372}

func (i SchemaKindType) String() string {
	if i >= SchemaKindType(len(_SchemaKindType_index)-1) {
		return "SchemaKindType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SchemaKindType_name[_SchemaKindType_index[i]:_SchemaKindType_index[i+1]]
}