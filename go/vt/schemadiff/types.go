/*
Copyright 2022 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schemadiff

import (
	"strings"

	"vitess.io/vitess/go/sqlescape"
	"vitess.io/vitess/go/vt/sqlparser"
)

type InstantDDLCapability int

const (
	InstantDDLCapabilityUnknown InstantDDLCapability = iota
	InstantDDLCapabilityIrrelevant
	InstantDDLCapabilityImpossible
	InstantDDLCapabilityPossible
)

// Entity stands for a database object we can diff:
// - A table
// - A view
type Entity interface {
	// Name of entity, ie table name, view name, etc.
	Name() string
	// Diff returns an entitty diff given another entity. The diff direction is from this entity and to the other entity.
	Diff(other Entity, hints *DiffHints) (diff EntityDiff, err error)
	// Create returns an entity diff that describes how to create this entity
	Create() EntityDiff
	// Drop returns an entity diff that describes how to drop this entity
	Drop() EntityDiff
	// Clone returns a deep copy of the entity.
	Clone() Entity
}

// EntityDiff represents the diff between two entities
type EntityDiff interface {
	// IsEmpty returns true when the two entities are considered identical
	IsEmpty() bool
	// EntityName returns the name of affected entity
	EntityName() string
	// Entities returns the two diffed entitied, aka "from" and "to"
	Entities() (from Entity, to Entity)
	// Statement returns a valid SQL statement that applies the diff, e.g. an ALTER TABLE ...
	// It returns nil if the diff is empty
	Statement() sqlparser.Statement
	// StatementString "stringifies" this diff's Statement(). It returns an empty string if the diff is empty
	StatementString() string
	// CanonicalStatementString "stringifies" this diff's Statement() to a canonical string. It returns an empty string if the diff is empty
	CanonicalStatementString() string
	// SubsequentDiff returns a followup diff to this one, if exists
	SubsequentDiff() EntityDiff
	// SetSubsequentDiff updates the existing subsequent diff to the given one
	SetSubsequentDiff(EntityDiff)
	// InstantDDLCapability returns the ability of this diff to run with ALGORITHM=INSTANT
	InstantDDLCapability() InstantDDLCapability
	// Clone returns a deep copy of the entity diff, and of all referenced entities.
	Clone() EntityDiff
	Annotated() (from *TextualAnnotations, to *TextualAnnotations, unified *TextualAnnotations)
}

const (
	AutoIncrementIgnore int = iota
	AutoIncrementApplyHigher
	AutoIncrementApplyAlways
)

const (
	RangeRotationFullSpec = iota
	RangeRotationDistinctStatements
	RangeRotationIgnore
)

const (
	ConstraintNamesIgnoreVitess = iota
	ConstraintNamesIgnoreAll
	ConstraintNamesStrict
)

const (
	ColumnRenameAssumeDifferent = iota
	ColumnRenameHeuristicStatement
)

const (
	TableRenameAssumeDifferent = iota
	TableRenameHeuristicStatement
)

const (
	FullTextKeyDistinctStatements = iota
	FullTextKeyUnifyStatements
)

const (
	TableCharsetCollateStrict int = iota
	TableCharsetCollateIgnoreEmpty
	TableCharsetCollateIgnoreAlways
)

const (
	TableQualifierDefault int = iota
	TableQualifierDeclared
)

const (
	AlterTableAlgorithmStrategyNone int = iota
	AlterTableAlgorithmStrategyInstant
	AlterTableAlgorithmStrategyInplace
	AlterTableAlgorithmStrategyCopy
)

const (
	EnumReorderStrategyAllow int = iota
	EnumReorderStrategyReject
)

const (
	ForeignKeyCheckStrategyStrict int = iota
	ForeignKeyCheckStrategyIgnore
)

const (
	SubsequentDiffStrategyAllow int = iota
	SubsequentDiffStrategyReject
)

// DiffHints is an assortment of rules for diffing entities
type DiffHints struct {
	StrictIndexOrdering         bool
	AutoIncrementStrategy       int
	RangeRotationStrategy       int
	ConstraintNamesStrategy     int
	ColumnRenameStrategy        int
	TableRenameStrategy         int
	FullTextKeyStrategy         int
	TableCharsetCollateStrategy int
	TableQualifierHint          int
	AlterTableAlgorithmStrategy int
	EnumReorderStrategy         int
	ForeignKeyCheckStrategy     int
	SubsequentDiffStrategy      int
}

func EmptyDiffHints() *DiffHints {
	return &DiffHints{}
}

const (
	ApplyDiffsNoConstraint = "ApplyDiffsNoConstraint"
	ApplyDiffsInOrder      = "ApplyDiffsInOrder"
	ApplyDiffsSequential   = "ApplyDiffsSequential"
)

type ForeignKeyTableColumns struct {
	Table   string
	Columns []string
}

func (f ForeignKeyTableColumns) Escaped() string {
	var b strings.Builder
	b.WriteString(sqlescape.EscapeID(f.Table))
	b.WriteString(" (")
	escapedColumns := make([]string, len(f.Columns))
	for i, column := range f.Columns {
		escapedColumns[i] = sqlescape.EscapeID(column)
	}
	b.WriteString(strings.Join(escapedColumns, ", "))
	b.WriteString(")")
	return b.String()
}
