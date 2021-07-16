package snowflakes

import (
	"testing"
)

var flakes = NewClientWithSigningKey("test_signing_key")

func TestRelationalFlakes(t *testing.T) {
	parent, _ := flakes.Gen("test_parent")
	child, _ := flakes.GenChild("test_child", parent)
	nestedChild, _ := flakes.GenChild("test_nested_child", child)

	recreatedChild, _ := flakes.GenParent(nestedChild, "test_child")
	if child != recreatedChild {
		t.Errorf("Expected %v got %v", parent, recreatedChild)
	}

	recreatedParent, _ := flakes.GenParent(recreatedChild, "test_parent")
	if parent != recreatedParent {
		t.Errorf("Expected %v got %v", parent, recreatedParent)
	}
}

func TestJSFlakes(t *testing.T) {
	parent := "test_parent_c6df4ae7a44744cc438fef30c091"
	child := "test_child_761cf03a0e7ff4fc7344ced40370f42f740a10fb26"
	nestedChild := "test_nested_child_c626af0f9a0ae7f704f4173764d48d4ce7732f4f6f7ff0a060f0f361"

	recreatedChild, _ := flakes.GenParent(nestedChild, "test_child")
	if child != recreatedChild {
		t.Errorf("Expected %v got %v", parent, recreatedChild)
	}

	recreatedParent, _ := flakes.GenParent(recreatedChild, "test_parent")
	if parent != recreatedParent {
		t.Errorf("Expected %v got %v", parent, recreatedParent)
	}
}

func TestElixirFlakes(t *testing.T) {
	parent := "test_parent_254fbb1e371e544567cf9f307040"
	child := "test_child_e50ef02b07eff7f6e7b4539417e6f7afeb0b00fb15"
	nestedChild := "test_nested_child_e5152f0f6b0b9efe67f7ee7e449419459be71f7fcfefc0b000f0d250"

	recreatedChild, _ := flakes.GenParent(nestedChild, "test_child")
	if child != recreatedChild {
		t.Errorf("Expected %v got %v", parent, recreatedChild)
	}

	recreatedParent, _ := flakes.GenParent(recreatedChild, "test_parent")
	if parent != recreatedParent {
		t.Errorf("Expected %v got %v", parent, recreatedParent)
	}
}
