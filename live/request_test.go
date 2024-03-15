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

package live

import (
	"net/http"
	"testing"
)

func TestClient_VerifyH5RequestSignatureWithParams(t *testing.T) {
	simpSDK := NewClient(&Config{
		AccessKeySecret: "NPRZADNURSKNGYDFMDKJOOTLQMGDHL",
	})

	url := "https://play-live.bilibili.com/plugins-full/1234567?Timestamp=1650012983&Code=460803&Mid=110000345&Caller=bilibili&CodeSign=8c1fa83955d83960680277122bd31fd6f209a82787d57912c1d3817487bfc2ef"
	testReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	h5sp := ParseH5SignatureParamsWithRequest(testReq)

	if !simpSDK.VerifyH5RequestSignatureWithParams(h5sp) {
		t.Fatal("VerifyH5RequestSignatureWithParams failed")
		return
	}

	t.Log("VerifyH5RequestSignatureWithParams success")
}
