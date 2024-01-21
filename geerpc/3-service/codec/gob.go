package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

// 这个是确保GobCodec实现了Codec接口，不然就是会出错
// var _ Codec = (*GobCodec)(nil)

//使用与客户端的socket连接初始化编解码器。
//dec: gob.NewDecoder(conn)使得解码时从连接中获取数据；
//enc: gob.NewEncoder(buf)编码器需要缓冲区，且该缓冲区的底层io流应该为与客户端的连接。
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *GobCodec) ReadBody(body any) error {
	return c.dec.Decode(body)
}

func (c *GobCodec) WriteResponse(h *Header, body any) (err error) {
	defer func() {
		c.buf.Flush()
		if err != nil {
			c.Close()
		}
	}()

	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}
