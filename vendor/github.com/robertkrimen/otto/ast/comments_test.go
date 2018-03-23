package ast

import (
	"github.com/robertkrimen/otto/file"
	"testing"
)

func TestCommentMap(t *testing.T) {
	statement := &EmptyStatement{file.Idx(1)}
	comment := &Comment{1, "test", LEADING}

	cm := CommentMap{}
	cm.AddComment(statement, comment)

	if cm.Size() != 1 {
		t.Errorf("the number of comments is %v, not 1", cm.Size())
	}

	if len(cm[statement]) != 1 {
		t.Errorf("the number of comments is %v, not 1", cm.Size())
	}

	if cm[statement][0].Text != "test" {
		t.Errorf("the text is %v, not \"test\"", cm[statement][0].Text)
	}
}

func TestCommentMap_move(t *testing.T) {
	statement1 := &EmptyStatement{file.Idx(1)}
	statement2 := &EmptyStatement{file.Idx(2)}
	comment := &Comment{1, "test", LEADING}

	cm := CommentMap{}
	cm.AddComment(statement1, comment)

	if cm.Size() != 1 {
		t.Errorf("the number of comments is %v, not 1", cm.Size())
	}

	if len(cm[statement1]) != 1 {
		t.Errorf("the number of comments is %v, not 1", cm.Size())
	}

	if len(cm[statement2]) != 0 {
		t.Errorf("the number of comments is %v, not 0", cm.Size())
	}

	cm.MoveComments(statement1, statement2, LEADING)

	if cm.Size() != 1 {
		t.Errorf("the number of comments is %v, not 1", cm.Size())
	}

	if len(cm[statement2]) != 1 {
		t.Errorf("the number of comments is %v, not 1", cm.Size())
	}

	if len(cm[statement1]) != 0 {
		t.Errorf("the number of comments is %v, not 0", cm.Size())
	}
}
