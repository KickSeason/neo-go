package stackitem

import (
	"errors"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/encoding/bigint"
	"github.com/nspcc-dev/neo-go/pkg/io"
)

// SerializeItem encodes given Item into the byte slice.
func SerializeItem(item Item) ([]byte, error) {
	w := io.NewBufBinWriter()
	EncodeBinaryStackItem(item, w.BinWriter)
	if w.Err != nil {
		return nil, w.Err
	}
	return w.Bytes(), nil
}

// EncodeBinaryStackItem encodes given Item into the given BinWriter. It's
// similar to io.Serializable's EncodeBinary, but works with Item
// interface.
func EncodeBinaryStackItem(item Item, w *io.BinWriter) {
	serializeItemTo(item, w, make(map[Item]bool))
}

func serializeItemTo(item Item, w *io.BinWriter, seen map[Item]bool) {
	if seen[item] {
		w.Err = errors.New("recursive structures can't be serialized")
		return
	}

	switch t := item.(type) {
	case *ByteArray:
		w.WriteBytes([]byte{byte(ByteArrayT)})
		w.WriteVarBytes(t.Value().([]byte))
	case *Buffer:
		w.WriteBytes([]byte{byte(BufferT)})
		w.WriteVarBytes(t.Value().([]byte))
	case *Bool:
		w.WriteBytes([]byte{byte(BooleanT)})
		w.WriteBool(t.Value().(bool))
	case *BigInteger:
		w.WriteBytes([]byte{byte(IntegerT)})
		w.WriteVarBytes(bigint.ToBytes(t.Value().(*big.Int)))
	case *Interop:
		w.Err = errors.New("interop item can't be serialized")
	case *Array, *Struct:
		seen[item] = true

		_, isArray := t.(*Array)
		if isArray {
			w.WriteBytes([]byte{byte(ArrayT)})
		} else {
			w.WriteBytes([]byte{byte(StructT)})
		}

		arr := t.Value().([]Item)
		w.WriteVarUint(uint64(len(arr)))
		for i := range arr {
			serializeItemTo(arr[i], w, seen)
		}
	case *Map:
		seen[item] = true

		w.WriteBytes([]byte{byte(MapT)})
		w.WriteVarUint(uint64(len(t.Value().([]MapElement))))
		for i := range t.Value().([]MapElement) {
			serializeItemTo(t.Value().([]MapElement)[i].Key, w, seen)
			serializeItemTo(t.Value().([]MapElement)[i].Value, w, seen)
		}
	}
}

// DeserializeItem decodes Item from the given byte slice.
func DeserializeItem(data []byte) (Item, error) {
	r := io.NewBinReaderFromBuf(data)
	item := DecodeBinaryStackItem(r)
	if r.Err != nil {
		return nil, r.Err
	}
	return item, nil
}

// DecodeBinaryStackItem decodes previously serialized Item from the given
// reader. It's similar to the io.Serializable's DecodeBinary(), but implemented
// as a function because Item itself is an interface. Caveat: always check
// reader's error value before using the returned Item.
func DecodeBinaryStackItem(r *io.BinReader) Item {
	var t = Type(r.ReadB())
	if r.Err != nil {
		return nil
	}

	switch t {
	case ByteArrayT:
		data := r.ReadVarBytes()
		return NewByteArray(data)
	case BooleanT:
		var b = r.ReadBool()
		return NewBool(b)
	case IntegerT:
		data := r.ReadVarBytes()
		num := bigint.FromBytes(data)
		return NewBigInteger(num)
	case ArrayT, StructT:
		size := int(r.ReadVarUint())
		arr := make([]Item, size)
		for i := 0; i < size; i++ {
			arr[i] = DecodeBinaryStackItem(r)
		}

		if t == ArrayT {
			return NewArray(arr)
		}
		return NewStruct(arr)
	case MapT:
		size := int(r.ReadVarUint())
		m := NewMap()
		for i := 0; i < size; i++ {
			key := DecodeBinaryStackItem(r)
			value := DecodeBinaryStackItem(r)
			if r.Err != nil {
				break
			}
			m.Add(key, value)
		}
		return m
	default:
		r.Err = errors.New("unknown type")
		return nil
	}
}
