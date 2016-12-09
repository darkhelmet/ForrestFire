package goredis

import (
	"strconv"
)

// Integer reply: the length of the string after the append operation.
func (r *Redis) Append(key, value string) (int64, error) {
	rp, err := r.ExecuteCommand("APPEND", key, value)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Count the number of set bits (population counting) in a string.
func (r *Redis) BitCount(key string, start, end int) (int64, error) {
	rp, err := r.ExecuteCommand("BITCOUNT", key, start, end)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Perform a bitwise operation between multiple keys (containing string values)
// and store the result in the destination key.
// The BITOP command supports four bitwise operations:
// AND, OR, XOR and NOT, thus the valid forms to call the command are:
// BITOP AND destkey srckey1 srckey2 srckey3 ... srckeyN
// BITOP OR destkey srckey1 srckey2 srckey3 ... srckeyN
// BITOP XOR destkey srckey1 srckey2 srckey3 ... srckeyN
// BITOP NOT destkey srckey
// Return value: Integer reply
// The size of the string stored in the destination key, that is equal to the size of the longest input string.
func (r *Redis) BitOp(operation, destkey string, keys ...string) (int64, error) {
	args := packArgs("BITOP", operation, destkey, keys)
	rp, err := r.ExecuteCommand(args...)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Decrements the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
// This operation is limited to 64 bit signed integers.
// Integer reply: the value of key after the decrement
func (r *Redis) Decr(key string) (int64, error) {
	rp, err := r.ExecuteCommand("DECR", key)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Decrements the number stored at key by decrement.
func (r *Redis) DecrBy(key string, decrement int) (int64, error) {
	rp, err := r.ExecuteCommand("DECRBY", key, decrement)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Get the value of key.
// If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string,
// because GET only handles string values.
func (r *Redis) Get(key string) ([]byte, error) {
	rp, err := r.ExecuteCommand("GET", key)
	if err != nil {
		return nil, err
	}
	return rp.BytesValue()
}

// Returns the bit value at offset in the string value stored at key.
// When offset is beyond the string length,
// the string is assumed to be a contiguous space with 0 bits.
// When key does not exist it is assumed to be an empty string,
// so offset is always out of range and the value is also assumed to be a contiguous space with 0 bits.
func (r *Redis) GetBit(key string, offset int) (int64, error) {
	rp, err := r.ExecuteCommand("GETBIT", key, offset)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Returns the substring of the string value stored at key,
// determined by the offsets start and end (both are inclusive).
// Negative offsets can be used in order to provide an offset starting from the end of the string.
// So -1 means the last character, -2 the penultimate and so forth.
// The function handles out of range requests by limiting the resulting range to the actual length of the string.
func (r *Redis) GetRange(key string, start, end int) (string, error) {
	rp, err := r.ExecuteCommand("GETRANGE", key, start, end)
	if err != nil {
		return "", err
	}
	return rp.StringValue()
}

// Atomically sets key to value and returns the old value stored at key.
// Returns an error when key exists but does not hold a string value.
func (r *Redis) GetSet(key, value string) ([]byte, error) {
	rp, err := r.ExecuteCommand("GETSET", key, value)
	if err != nil {
		return nil, err
	}
	return rp.BytesValue()
}

// Increments the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
// Integer reply: the value of key after the increment
func (r *Redis) Incr(key string) (int64, error) {
	rp, err := r.ExecuteCommand("INCR", key)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Increments the number stored at key by increment.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
// Integer reply: the value of key after the increment
func (r *Redis) IncrBy(key string, increment int) (int64, error) {
	rp, err := r.ExecuteCommand("INCRBY", key, increment)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Bulk reply: the value of key after the increment.
func (r *Redis) IncrByFloat(key string, increment float64) (float64, error) {
	rp, err := r.ExecuteCommand("INCRBYFLOAT", key, increment)
	if err != nil {
		return 0.0, err
	}
	s, err := rp.StringValue()
	if err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(s, 64)
}

// Returns the values of all specified keys.
// For every key that does not hold a string value or does not exist,
// the special value nil is returned. Because of this, the operation never fails.
// Multi-bulk reply: list of values at the specified keys.
func (r *Redis) MGet(keys ...string) ([][]byte, error) {
	args := packArgs("MGET", keys)
	rp, err := r.ExecuteCommand(args...)
	if err != nil {
		return nil, err
	}
	return rp.BytesArrayValue()
}

// Sets the given keys to their respective values.
// MSET replaces existing values with new values, just as regular SET.
// See MSETNX if you don't want to overwrite existing values.
func (r *Redis) MSet(pairs map[string]string) error {
	args := packArgs("MSET", pairs)
	_, err := r.ExecuteCommand(args...)
	return err
}

// Sets the given keys to their respective values.
// MSETNX will not perform any operation at all even if just a single key already exists.
// True if the all the keys were set.
// False if no key was set (at least one key already existed).
func (r *Redis) MSetnx(pairs map[string]string) (bool, error) {
	args := packArgs("MSETNX", pairs)
	rp, err := r.ExecuteCommand(args...)
	if err != nil {
		return false, err
	}
	return rp.BoolValue()
}

// PSETEX works exactly like SETEX with the sole difference that
// the expire time is specified in milliseconds instead of seconds.
func (r *Redis) PSetex(key string, milliseconds int, value string) error {
	_, err := r.ExecuteCommand("PSETEX", key, milliseconds, value)
	return err
}

// Set key to hold the string value.
// If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation.
func (r *Redis) Set(key, value string, seconds, milliseconds int, must_exists, must_not_exists bool) error {
	args := packArgs("SET", key, value)
	if seconds > 0 {
		args = append(args, "EX", seconds)
	}
	if milliseconds > 0 {
		args = append(args, "PX", milliseconds)
	}
	if must_exists {
		args = append(args, "XX")
	} else if must_not_exists {
		args = append(args, "NX")
	}
	rp, err := r.ExecuteCommand(args...)
	if err != nil {
		return err
	}
	return rp.OKValue()
}

// Sets or clears the bit at offset in the string value stored at key.
// Integer reply: the original bit value stored at offset.
func (r *Redis) SetBit(key string, offset, value int) (int64, error) {
	rp, err := r.ExecuteCommand("SETBIT", key, offset, value)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Set key to hold the string value and set key to timeout after a given number of seconds.
func (r *Redis) Setex(key string, seconds int, value string) error {
	rp, err := r.ExecuteCommand("SETEX", key, seconds, value)
	if err != nil {
		return err
	}
	return rp.OKValue()
}

// Set key to hold string value if key does not exist.
func (r *Redis) Setnx(key, value string) (bool, error) {
	rp, err := r.ExecuteCommand("SETNX", key, value)
	if err != nil {
		return false, err
	}
	return rp.BoolValue()
}

// Overwrites part of the string stored at key, starting at the specified offset, for the entire length of value.
// Integer reply: the length of the string after it was modified by the command.
func (r *Redis) SetRange(key string, offset int, value string) (int64, error) {
	rp, err := r.ExecuteCommand("SETRANGE", key, offset, value)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}

// Returns the length of the string value stored at key.
// An error is returned when key holds a non-string value.
// Integer reply: the length of the string at key, or 0 when key does not exist.
func (r *Redis) StrLen(key string) (int64, error) {
	rp, err := r.ExecuteCommand("STRLEN", key)
	if err != nil {
		return 0, err
	}
	return rp.IntegerValue()
}
