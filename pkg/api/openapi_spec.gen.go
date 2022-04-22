// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+R92W4cOZborxDRF3A3buSixZuertsud8m3yvZYctcAZUNiRJzIpBVJZpEMyVmGgP6I",
	"+ZOZBuZh+ml+wP1HA55DxpLBlFK25XL31IORyowgD8++sj4kuVoslQRpTXLwITH5HBYcPz4yRswkFMfc",
	"nLm/CzC5FksrlEwOer8yYRhn1n3ihgnr/taQgziHgmUrZufAflL6DPQ4SZOlVkvQVgDukqvFgssCPwsL",
	"C/zwfzSUyUHyu0kL3MRDNnlMLySXaWJXS0gOEq41X7m/36nMve2/NlYLOfPfnyy1UFrYVecBIS3MQIcn",
	"6NvI65Iv4j9cvaax3NbXHsfh74iedCfi5mwzIHUtCvdDqfSC2+SAvkjXH7xMEw2/1EJDkRz8HB5yyPFn",
	"aWDrHGENSx2UdKFKW3q9bfZV2TvIrQPw0TkXFc8qeKayI7DWgTPgnCMhZxUwQ78zVTLOnqmMudVMhEHm",
	"SuT0sb/OT3OQbCbOQaasEgthkc/OeSUK928NhlnlvjPA/CJj9kJWK1YbByO7EHbOCGm4udu7YcEB8teZ",
	"rYCS15UdwnU8B+Z/JDiYmasL6YFhtQHNLhzsBVjQCyFx/7kwASVjWr6zZnyL5puJVaqyYuk3ErLdyPGj",
	"LnkOuCgUwrqj04oe/pJXBtIhcu0ctAOaV5W6YO7VdUAZL617Zg7sncrYnBuWAUhm6mwhrIVizH5SdVUw",
	"sVhWK1ZABfRaVTF4LwwtyM2ZYaXStPQ7laWMy8IpELVYiso9I+z4jWwZPVOqAi7xROe8GuLn5crOlWTw",
	"fqnBGKEQ+Rkw93TNLRQOR0oXdMBAB8CT9EnXwNXQJh2yxhmshjAcFiCtKAVov0jD8ilb1MY6eGopfqmJ",
	"ET3R3nlBiO7jBIPrWUQWHskVg/dWc8b1rF44DRP4LVuuxu5FMz5SC3hJsrX6/R9Y7shQGyjck7kGboGO",
	"6uVv1YGhFfFWs9yAhcRiAYXgFqoV0+CWYhyPWkAppHAvpE4R4PZuyxRxomrrIeLairyuuG7osIEfTJ0F",
	"9XmV1o0oqiP/ZiPqN17h2L9+LoxYFzKr66sQ5AS3L1qeH14fkoJ0yApipdnvK3EGjLM/ViAdE/OiGCn5",
	"hzE7AuuWO0WCnJKaIXvMJekCyatmDzvn1m1dV4W8gwzZaCqQBSoQE0f0molxAuAf2tIsHLV0WrMOdTZy",
	"vxA7kEAEmrPHtdYgbbViyulxHtZFCetocjNmp98/Ovr+uycnTw9/+O7k5aPj70/JSymEhtwqvWJLbufs",
	"/7LTN8nkd/jfm+SU8eXSobSgY4OsF+58pajgxD2fpEkhdPiIX3uLOudmDsVJ++TbiABvYpqhgvcY6Jy+",
	"ozXIfHHDDp8EecZjO6bxLDFmzxWTYJyuM1bXua01GPZ7NF8mZYXI3VZcCzB/YFwDM/VyqbRdP7oHPnWe",
	"zd6uO3SluE1S5IVrDxk/XbD27Z7kJQrDfuSSz0CTCRAWRZ8vnIKOuAYVz6C6mcvmkbm9uxlzaQbewJo4",
	"eJYg8Dp7XicbDlsR5f6DMDYwA3L3ZrwNcRTcuE878XFPI244brtF7IDBXx8cy//ANDgrjSaLM0POofcy",
	"URO9h7y2cF0csdlJbxio83MAL064ziuxE32ntdLD8/wJJGiRM3A/Mw1mqaSBWMRTRGTi++Pjl4zccuae",
	"aMxhsxA7NEzIvKoL8l8cNpZ8VSleMKNImTcIJGh7uHVOF4ImJAUQQsnxG/nYbXZ3uufUI3oDaDTQE+KW",
	"Z9yA+yWrzWrMnN+JgAag2IWoKpYrabmQjLM7r8Dq1eiR8wvv0KNz4OhnOfCELETOLRjvOV7MRT5nVizI",
	"9XKkAGNZzqXTcxqsFs6JfKqcC0oWC8KCwjCpLHNswp0+DyrjjmH1Mjg/eSVAokksFDNqAc7RmjEN3CiJ",
	"WhQtKbwnIRC8YhnPz1RZkhZsIq2g/YZh3gKM4bMY760xF9K9fT7GWU8rvgCZqz+DNt7x35LLz9s3roYi",
	"POh1ZAyKZxRG86p6USYHP1+tLY6Cr+/eukzXAea5Fec+RO4z/JP2r2DrK24sC28w57X7iCDqDZPLGlMs",
	"7gd0+sUCjOWLZZeSBbcwcr/E1hSR5V6/PnwSIHyGQfQ18fe2ob+zKE3kXy+L+GmOwyEcDIghenS85aHW",
	"6I8AB9S123ZSAg3J3l6+JW74ESx3ygAJWhTor/PqZY/QAxysBSg6E1ZzvWILv5j318yY/ag0avxlBe+7",
	"zopXAwvlAkc0dbXTbuyUj7NxfurEn+gc4qszwJgF3nO3lpce5OqD5GiphQX2VIvZ3Lkvzrkdw4KLykG9",
	"yjTI/5d530npWXiCBC45wgfYkf3v/zqHqmMRe1Jz1HEu4niiMCD6bsMywfNCOmDKhcvcYYCyL8sKrP8s",
	"CVlCyVHJBT3RfFhyp82TNPmlhho/cJ3PxXnnIzl2tPzIK99mkd4X+JlWqR2KRt3NkzS54JgdGJVKj5wL",
	"bKKeYXPMx3MuZzBUb6SX4/kO+q0TkHtbiUuNv4gUrolKIxEerA268pibM3NULxZcr2LZrsWyEqWAglXe",
	"j6OMRwjrxuwxmU8y0fhjyrLaonlzXzl77R4H7owlN2dDnwLf2trDw5yjB3gL585sOvlrVB7RGM6AbTJ7",
	"KXMeCFPnoNmRys/AHr4gL4PiWiKhcXZeMwkX7kuTstOlhnOhanNChDglZyNz5pq8ILLRfUx8IQ0eDG1/",
	"oed80Q1M4zmaHtA3sgHdfHKTObg7TT85udxbfXNa+aam5zMsT8Te+MRyc/gNzGb+pQYSsI5+x7xvcnDX",
	"uVWtjdqk9S/TBJN+J9kKE+PrsLwNn06E7GngRvl57fr2chCKEiAfkoWQYuEU+E7cWfxsS/pUVM51zlpL",
	"mga7+MPh//+uNYvR9J0qSwN9QKP81eLpww1y4mZLA7jpRJ10h7nJqTpUW+fhV2BrLSlf5PQKZf15MB/C",
	"O5l4hF4B4IZy1VGfm7n3FRhfMhgE6dtrb3K0P1Fr+zzBYyVLMas1t9Eww8z5gsvvMEIqopUXSlzOgR3h",
	"o8wZfWY1l6YEzR69PMRsW8gkjOO5Wqs0n8EPKufxMseTJleHgakz/Y5DcC//8vhatbO+S7p2uhiWXsFM",
	"GAsaCko3DDHEi0KDiUuF05Qn3cBnaF1EfrY5YVFx69RrPH+lSnvB9Ybk1lZGgY7U8m+TTDpp6nzmZmL/",
	"WXXJBhdpg9RufTIgI01ySv4ilMk6ljuY2XCiGJ2PIK+dzWkyOn0ibx3aXxXTk4A8nkN+pupIufCIPCX0",
	"mkk52TkIzY6+f7R79x7L3YumXqTMiF8xw5utLBjKdhRgHAis8swd0kK5363Ndq/FReSju9gec9UHSVuI",
	"Gc8UyUhykOzdzab7D3fy3fvZdG9vr9gps/27ZT69/+Ah39nN+fRetlPc258Wu3fvPbz/YJo9mN4v4O50",
	"v7g/3X0IU7eQ+BWSg5393X1MDtBulZrNhJx1t7q3l93fze/tZQ/3d/fLYmcve7h3f1pm96bTew+nD6b5",
	"Ht+5e3/nfl7u8WJ/f/fe3t1s58H9/B5/8PDu9P7Ddqvd+5dD+xww8hIBGNQLuZ07j1RTHsorSV8a6dXG",
	"wjpjdujbHCrunISQcPLqsCEAFhm4YblXuFBQXqPZZMwOJVNVAZr51IwJHqZfC/e94Ia9qw3VuN80x2GH",
	"T94kFC0ES+ZXYaLJo3GCAjNdp943Gpmqnk1MDhJGTvomVIocHT457VV8WqH3LLOlkSLYn4oKjpaQX2uv",
	"aPG0T6brpam1p7F40f1GQcgaVWJNBp/AHj6Lss4Yx/gnob4QZQkaU5BzLtnFnFskZRNbp445uotijAPS",
	"1NoRzheIWzHGFCyS84swX4zU62nL7UjSkHqo4JaQi1J4DYX0QAvudZUHumPP+6RZRkkSzHmQle6KAeJo",
	"TmDOIxD2VW13zegaqGc+DL1Y6OvoSLp43TeZ86C30mS5HYJ/EnbeJou2QnXqE+05qrNsA+pT5sJvZVNW",
	"wBJkgc05EotwZI7/yWmzrf/UIceGvNGAqt0I8yryDnKAtTyT6kJi4FwpXlDGzhGs57m256fFXhE02Afy",
	"ilTNJzse6Gj0cLfRl7glp+GrOAhfwbxtJn6fXlRai1s1olap1YJxpjuvBZOSdknpg1zVF3fQ587veIpL",
	"UdlQA0NGc5bEP+a+C4k22pBKdG1Z82vxQCuYjTzcDlt0N2rE7QvzSkd9fy7XUCNlX3Gsibin/01t7pdS",
	"hFcovW7ZMNog0kYkbT+hY89QI13jwG3yf59f9fE/7H38N/b3v3z868e/ffyPj3/9+18+/ufHv338924S",
	"E9O23XSY3+UkXxTJQfLB/3mJPm8tz06ICffcmazmuT3hdSFUSJg54vnYaaLxzYkpJ+9UZsiH39ndG+OS",
	"3dTsy+d/cn8uTXLghKjUfOHIm+yMdpyAiQWfgTlR+uRcFKCcKcRvkjRRtV3WlpqZ4L0FSbXlZLxE+0MQ",
	"nPinhnDRTg1kkzi6fNfVYD2tlL1yvY7gGOHoP/LYHNEryUBgu8xxTWqtqXVu2/h9XaK+wwPXpRTDo5uT",
	"9fF+lW0y6J0+8xvUGJtqYpMBN6q0bbUxUjv0dceYc9KtOw1l3kW3IFU9m3c7VBjPqA0UaMAgNEu2/bJ3",
	"DP1SCROJ7LbKTH1LPf5tMeZTayxX9Ph3O/nDTpu4ZVN1r/2NYb+qtCxbMe47sxyBaGVqeSYL9KaeTnfv",
	"sUrNvDVCimGzDvV3+dbqbZtHXkgYVUL67mLfVYTh9R3D8qZLdI7tnC6QCU4R1U/Zi3PQF840GBYKdtWK",
	"ztL0n4T+hBi7VGoWi3tmzAHV6WZ3u6VN3TI0lzqgERW4IXBdCWppG2Zae1K7LY/FahBEHUpsb0r7f0Za",
	"GnJNZaXhT5+ZXl73KWinXmY4ukUns/x2Iz6OxEy+uCkmQqb5ZHPT1Rc/didLvuG0A6iuOLXlFjb1YPha",
	"WqvSb1ROiHqAncW2AqrYBNUXgOUaCPrm0ViuLcXc/IKfYY3CVABL53tizSBNzLy2BcXoFox/WpWl0wQR",
	"K0jCglWHIwc1He8CATjhdSyf8tqAdrR36tapMHqYHT5J2ZIbc6F0EX4i6aCpKMZteFR3xN7pGcQXVuO4",
	"EXmreObWLpNLB6OzwNSsKi3Pbdt72PQosmPgTvhqXfk3zcFkUgbvXKjJsHb9isYVnnK9YAuf8Hz08jBJ",
	"k0rk4INev8+fXv5wvjdY/+LiYjyTtXPWJ/4dM5ktq9HeeDoGOZ7bBbVnCVv1oPXbJZ1WyWRnPB1Psdq9",
	"BMmXwnn2+BWlbZAyE74Uk3y9XDojZec4FL87LLAB2Pbrqo7/KFzGpXan04BSkPg+Xy4rn62bvPOdT8TL",
	"13F6tI6LlOtjXDofomrCduK/4II5iCkr112m6Tzu9JRb7jzYn9GRxu6Hdo3vZLFUQlo0ejM/WTJYsKFD",
	"s+hlSrgNte+lMhGcUpxITTtei/xRFasvhsd+++oQfziqoHwEmnQVigvMLm+RwlcAdMENM3Wegynrqlox",
	"GpTDrnDvDp2Loua+mWy8Nq34RaCjimkEPvyBhYJon90I2YyHJi9kmXXO6PT8dzmPOhp6yz0LM1E04gee",
	"EfusNfkltA3FGQz7Mp65xW+HwdrOpQiyBvl6ytNjn4pVTprGX5vneo0qEZCfk0JBrDZqJQ31Llgs7Yoa",
	"HUXJpKJ2+wW3+RwLZUAvfjss+RRsPm86Mx3ir2G6FxmOOrStRCV2L+FYqyyYUboZ4W150JnXyQf373O+",
	"gMurLEgYuukPrvz8IRHuKL6a4E1kWHDAI2kHZev+x9tb5J/h6NAGjUq/rZsiP7sS5pw2zIBdQZxDWSqf",
	"MODMeOHqzNQOiGK2IIVJviLGTAxlzUPtCFgEe9VgTAwnqLAQsDUG260avfyunc3v4e/DO5WdiGIzN6Ns",
	"kem+npdpsSs5+br2ore/jTVGVzmmVUTLjT5VtYVBo5dk4R3WhYM8ivaJAdtGRRu8J2TioyYR9NWocCt2",
	"tDc7ECHGcZtvopZyZ0M9LNvY0f2NBRS/nPO7eJ7D0kKBwrC/u7up+cQHvGsA+aE1uqQijK34fFVTrStb",
	"dvmaZvK1hPdLyB3QGNuOKZG0mV19ZbltpvWHDOeiKDScI8LBzdTCleoD5yv+SXRIb1YkQgO0gvizgIYT",
	"mnmRdhzut+eLdWXHPdwrzO+GQZZwhA4rXG1+XFBjNpwYOaiTc9tkt//cjDjeGinXBzU/OfJurGxo21sL",
	"vq+OvR/jeCt1m9XGtyhaRb0b9JdwKsvW3MWJvN3ONwE0aKVE0UT7xufRRdv3HLUsoUPa90ffjs6P5Kwj",
	"iG7rDgH6rxozDXrFt+GFryi+9Zr4rjFiAN9H5ReBnoHr/BdvIy+12Z32TbPOUUbM5EiV5RVOipjJF2WZ",
	"bGOPvz1E+lwumqReFvfnt86YtDj7keuzbvqWOwVHWeJrsP2YV37+lTgMRbzyCiRkTc4k3iADqzsa2EzR",
	"zVq4/DhOEnkNReStCrXfYrM4NyX7rynLw/LIP4Qwb82Dj2o7B2mpWuprso4bQjX9orlc4wszpAZerNxT",
	"bj2aXejViUVL8CG7Wl+Gjtr7DsmS35ozENLgBbc1r8t0kzJjm9/4tlnq5uxBLslFOzmmga6lWm1AQpwP",
	"RnmnQhhVXpFq4q0qsu5GsXxpYxrpnJ8Wh/4D6xyvzz3dCAlhSCL0o2Nc7RRGBQXlPKhLxOuSUT+cDryC",
	"DetCtpfReP0CelSpnFeo2nhlvrQ+O4feaWozYFXrr0zdYF7zORR1Bcc0yHZ7ucXuBa6xfInvp2qqLZsU",
	"1XPlI7L+hWsYX4T7mC7TZH+69+Xqcr3JvAjwL0GHws8TkIKU5v70YWRalRjQp168paPGN2KnlBkVfsbL",
	"LqF38RQdHTtImVQXPvGz93VNS5AiLh2UiooQzu3u3BeB98PNFN7ZKRXqWZK2G0qsL3HwZv0ONq4TJeQp",
	"4xlcR2pyHQmZfMAGFp9CjstKpxFtq4oILfgtJjA7J9kki94f6nQ/fpq1OJ5DWGuYsoyJyHHornQW2WuN",
	"LhsR0VK6Zqy3NspMd/1/FLP0uu1RpCY9u1qKHNMk3ZbCpVYzDcak/oIif1WpZiUXVa3hWtsSLIoBWfQq",
	"Ag7dYXWnxZxHRGJC8+mTMCo1oTnDK+xJf8L4lhol+pvEitndeaLG4/Pjll8vhotOiEbADU8gG4dRzk5X",
	"RVdabpeTG0h4RXES3o1svKHZv30AjtEbv3D/EPXQssrZmL02wE7NGkbboaNTR2caLWWISmxfUBLM+FvK",
	"cT2mAe7O5a8UgprVohLyrLl7EGfpCQPUSWNp3tYjxZlXXlVszs+BLrqmKSHSlX6mJoMSrzPjVdVcl91a",
	"wVZZEFLXlMWRB4gz0xUmBKY318818Liy6M6EbasyuiS9VfURm0vcVpP8BkokOpYXg7e57wrv5FQYqXQJ",
	"kQaDEu6w9HNsdMRvS1Zw7LOdme/iwA8T+0tblbbGSzxRiuvmYNdy+iPnZ7tt2hs6Q4agv2AbcvgpRqpc",
	"EBStvqErjK2oqhaEjnjgepMPYab1cvIBvxG/XtGM0x1vUxoeeyZcc0K3nlbGy4WGHmt49EY9POnwlrlf",
	"YX3cupnVjewaTr/Nru3w+ttbl7jBSOPmDrR2EvVbk57u4Es7ehkdwqX7F4aCcpXWbjjyfzczprEgxmuT",
	"9p5hunaYrkIpoATNmsless2IDbTyb5Ld6YM3ydrdxhhuy2rlLySutexekUzHM43nRv24zSj1gOAUqPPK",
	"KFrDqAUoCQwquma5HV2KgYncggikO5BbFP7riLYZPeZy9MSdc/QaF0giOOzc4h/DodJiJiSvcE+3/pgd",
	"ln42qlLdWapm5FzYZsZp/YZqOjeOOzXXUHDJuMAnCshqugpoi7O98ICNnnrAkmv7HLdxZFRuwY6M1cAX",
	"fQ3RZAoyIZ18D3MFQ1+e9jBr91R8YhCP7DUI4XenD6573LNjjxE7Jf/9nfvRFbR/3QUA2LTLMrAX4Jk9",
	"XM/dKp3QYehbDPzNbCj+eqB3Gmc58DKGN3cj9zeTEPvbl66R2iCBreSEq9G1wl4SVbIM3IvN/tmqJ3fk",
	"SpxuFKED5mh2Sl33pF266PAn+VYsEFoGn7vbbHfYc4XJD26HP6J8lkrnIqtWLK+UoTQJ3uaeKykBb/31",
	"N7j6DJFXvKWQwszB9OgFDN7z3DLDF+BdSKtw7tK9UqjaeXf0ghm/kYGqd/DiIpImzwsZxCjAMlWsNprS",
	"bsoH78pvwoohWnwOyX0mg0rDR5OkU/Ma/H9++l2eg3EGYQ1U5bjVZ9jHM1S9z1QWSrKYG/qlBi3ApJ0R",
	"h3StMXTca30zkUUfvTzsD1l0K3Jqsailn5x1Kn04o9Ms71NbEVtP+Hv08jDFjZDlWuL7A2F6xf1N1/1S",
	"1Gk663t6Xb69/J8AAAD//wGTUz2RbgAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
