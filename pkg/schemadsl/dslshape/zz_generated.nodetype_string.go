// Code generated by "stringer -type=NodeType -output zz_generated.nodetype_string.go"; DO NOT EDIT.

package dslshape

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NodeTypeError-0]
	_ = x[NodeTypeFile-1]
	_ = x[NodeTypeComment-2]
	_ = x[NodeTypeDefinition-3]
	_ = x[NodeTypeRelation-4]
	_ = x[NodeTypePermission-5]
	_ = x[NodeTypeTypeReference-6]
	_ = x[NodeTypeSpecificTypeReference-7]
	_ = x[NodeTypeUnionExpression-8]
	_ = x[NodeTypeIntersectExpression-9]
	_ = x[NodeTypeExclusionExpression-10]
	_ = x[NodeTypeArrowExpression-11]
	_ = x[NodeTypeIdentifier-12]
	_ = x[NodeTypeNilExpression-13]
}

const _NodeType_name = "NodeTypeErrorNodeTypeFileNodeTypeCommentNodeTypeDefinitionNodeTypeRelationNodeTypePermissionNodeTypeTypeReferenceNodeTypeSpecificTypeReferenceNodeTypeUnionExpressionNodeTypeIntersectExpressionNodeTypeExclusionExpressionNodeTypeArrowExpressionNodeTypeIdentifierNodeTypeNilExpression"

var _NodeType_index = [...]uint16{0, 13, 25, 40, 58, 74, 92, 113, 142, 165, 192, 219, 242, 260, 281}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return "NodeType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
