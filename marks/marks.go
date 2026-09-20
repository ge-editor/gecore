/*
Mark C-space

Mark C-space は leaf の種類を意識せず、すべての leaf の Mark を
共通の interface として管理・参照する。

各 leaf は、それぞれの用途に応じた Mark を持つことができる。

editorleaf の Mark では、例えば以下の情報を持つ。

  - edit_buffer
  - cursor
  - label

これにより、Mark への移動だけでなく、editor 固有の region 操作なども
Mark を起点として行える。

他の leaf も、それぞれ必要な情報を持った独自の Mark を実装できる。

Mark は leaf をまたいで利用できる。

  - どの leaf からでも Mark を登録できる
  - どの leaf からでも Mark を削除できる
  - どの leaf からでも、別の leaf の Mark へ移動できる

C-space は個々の leaf の実装詳細を持たず、共通の Mark interface
だけを扱う。

Mark の状態変更は、主に各 leaf 側で行う。

例えば editorleaf では、カーソル位置を調整した際に Mark の位置や
label を変更するなど、leaf 自身が Mark の整合性を維持する。

つまり Mark は leaf 間を自由に移動するための共通の参照点となる。
*/

package marks

var Marks = NewMarks()

type Mark interface {
	// Leaf() tree.Leaf
	// LeafName() string

	SetLabel(string)
	GetLabel() string
}

// --------------
// Marks
// --------------

func NewMarks() *MarksStruct {
	return &MarksStruct{
		marks: make([]Mark, 0),
	}
}

type MarksStruct struct {
	marks []Mark
}

func (m *MarksStruct) Remove(target Mark) bool {
	for i, v := range m.marks {
		if v != target {
			continue
		}

		copy(m.marks[i:], m.marks[i+1:])
		m.marks[len(m.marks)-1] = nil
		m.marks = m.marks[:len(m.marks)-1]
		return true
	}

	return false
}

func (m *MarksStruct) Add(newMark Mark) Mark {
	m.marks = append([]Mark{newMark}, m.marks...)
	return newMark
}

func (m *MarksStruct) Items() []Mark {
	result := make([]Mark, len(m.marks))
	copy(result, m.marks)
	return result
}
