/*
 * MIT License
 *
 * Copyright (c) 2024 VTB-LINK and runstp.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS," WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE, AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
 * OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES, OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT, OR OTHERWISE, ARISING FROM, OUT OF,
 * OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package basic

import "sync"

type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, val []byte) error
	Del(key string) error
}

type MapStorage struct {
	m sync.Map
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		m: sync.Map{},
	}
}

func (s *MapStorage) Get(key string) ([]byte, error) {
	if val, ok := s.m.Load(key); ok {
		return val.([]byte), nil
	}

	return nil, nil
}

func (s *MapStorage) Set(key string, val []byte) error {
	s.m.Store(key, val)
	return nil
}

func (s *MapStorage) Del(key string) error {
	s.m.Delete(key)
	return nil
}
