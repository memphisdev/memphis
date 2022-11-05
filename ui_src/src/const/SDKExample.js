// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

export const CODE_EXAMPLE = {
    'Node.js': {
        langCode: 'js',
        installation: `npm i memphis-dev --save`,
        producer: `const memphis = require("memphis-dev");

        (async function () {
            let memphisConnection
        
            try {
                memphisConnection = await memphis.connect({
                    host: '<memphis-host>',
                    username: '<application type username>',
                    connectionToken: '<broker-token>'
                });
        
                const producer = await memphisConnection.producer({
                    stationName: '<station-name>',
                    producerName: '<producer-name>'
                });
        
                const headers = memphis.headers()
                headers.add('<key>', '<value>')
                await producer.produce({
                    message: Buffer.from("Message: Hello world"),
                    headers: headers
                });
        
                memphisConnection.close();
            } catch (ex) {
                console.log(ex);
                if (memphisConnection) memphisConnection.close();
            }
        })();
        `,
        consumer: `const memphis = require('memphis-dev');

        (async function () {
            let memphisConnection;
        
            try {
                memphisConnection = await memphis.connect({
                    host: '<memphis-host>',
                    username: '<application type username>',
                    connectionToken: '<broker-token>'
                });
        
                const consumer = await memphisConnection.consumer({
                    stationName: '<station-name>',
                    consumerName: '<consumer-name>',
                    consumerGroup: ''
                });
        
                consumer.on('message', (message) => {
                    console.log(message.getData().toString());
                    message.ack();
                    const headers = message.getHeaders()
                });
        
                consumer.on('error', (error) => {});
            } catch (ex) {
                console.log(ex);
                if (memphisConnection) memphisConnection.close();
            }
        })();
        `
    },

    Typescript: {
        langCode: 'js',
        installation: `npm i memphis-dev --save`,
        producer: `import memphis from 'memphis-dev';
        import type { Memphis } from 'memphis-dev/types';
        
        (async function () {
            let memphisConnection: Memphis;
        
            try {
                memphisConnection = await memphis.connect({
                    host: '<memphis-host>',
                    username: '<application type username>',
                    connectionToken: '<broker-token>'
                });
        
                const producer = await memphisConnection.producer({
                    stationName: '<station-name>',
                    producerName: '<producer-name>'
                });
        
                    const headers = memphis.headers()
                    headers.add('<key>', '<value>');
                    await producer.produce({
                        message: Buffer.from("Message: Hello world"),
                        headers: headers
                    });
        
                memphisConnection.close();
            } catch (ex) {
                console.log(ex);
                if (memphisConnection) memphisConnection.close();
            }
        })();
        `,
        consumer: `import memphis from 'memphis-dev';
        import { Memphis, Message } from 'memphis-dev/types';
        
        (async function () {
            let memphisConnection: Memphis;
        
            try {
                memphisConnection = await memphis.connect({
                    host: '<memphis-host>',
                    username: '<application type username>',
                    connectionToken: '<broker-token>'
                });
        
                const consumer = await memphisConnection.consumer({
                    stationName: '<station-name>',
                    consumerName: '<consumer-name>',
                    consumerGroup: ''
                });
        
                consumer.on('message', (message: Message) => {
                    console.log(message.getData().toString());
                    message.ack();
                    const headers = message.getHeaders()
                });
        
                consumer.on('error', (error) => {
                    console.log(error);
                });
            } catch (ex) {
                console.log(ex);
                if (memphisConnection) memphisConnection.close();
            }
        })();
        `
    },

    Go: {
        langCode: 'go',
        installation: `go get github.com/memphisdev/memphis.go`,
        producer: `package main

        import (
            "fmt"
            "os"
        
            "github.com/memphisdev/memphis.go"
        )
        
        func main() {
            conn, err := memphis.Connect("<memphis-host>", "<application type username>", "<broker-token>")
            if err != nil {
                os.Exit(1)
            }
            defer conn.Close()
            p, err := conn.CreateProducer("<station-name>", "<producer-name>")
        
            hdrs := memphis.Headers{}
            hdrs.New()
            err = hdrs.Add("<key>", "<value>")
        
            if err != nil {
                fmt.Errorf("Header failed: %v", err)
                os.Exit(1)
            }
        
            err = p.Produce([]byte("You have a message!"), memphis.MsgHeaders(hdrs))
        
            if err != nil {
                fmt.Errorf("Produce failed: %v", err)
                os.Exit(1)
            }
        }
        `,
        consumer: `package main

        import (
            "fmt"
            "os"
            "time"
        
            "github.com/memphisdev/memphis.go"
        )
        
        func main() {
            conn, err := memphis.Connect("<memphis-host>", "<application type username>", "<broker-token>")
            if err != nil {
                os.Exit(1)
            }
            defer conn.Close()
        
            consumer, err := conn.CreateConsumer("<station-name>", "<consumer-name>", memphis.PullInterval(15*time.Second))
        
            if err != nil {
                fmt.Printf("Consumer creation failed: %v\n", err)
                os.Exit(1)
            }
        
            handler := func(msgs []*memphis.Msg, err error) {
                if err != nil {
                    fmt.Printf("Fetch failed: %v\n", err)
                    return
                }
        
                for _, msg := range msgs {
                    fmt.Println(string(msg.Data()))
                    msg.Ack()
                    headers := msg.GetHeaders()
                    fmt.Println(headers)
                }
            }
        
            consumer.Consume(handler)
        
            // The program will close the connection after 30 seconds,
            // the message handler may be called after the connection closed
            // so the handler may receive a timeout error
            time.Sleep(30 * time.Second)
        }
        `
    },

    Python: {
        langCode: 'python',
        installation: `pip3 install memphis-py`,
        producer: `import asyncio
        from memphis import Memphis, Headers
        
        
        async def main():
            try:
                memphis = Memphis()
                await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<broker-token>")
        
                producer = await memphis.producer(
                    station_name="<station-name>", producer_name="<producer-name>")
                headers = Headers()
                headers.add("<key>", "<value>") 
                for i in range(5):
                    await producer.produce(bytearray('Message #'+str(i)+': Hello world', 'utf-8'), headers=headers)
        
            except Exception as e:
                print(e)
        
            finally:
                await memphis.close()
        
        if __name__ == '__main__':
            asyncio.run(main())
        `,
        consumer: `import asyncio
        from memphis import Memphis
        
        
        async def main():
            async def msg_handler(msgs, error):
                try:
                    for msg in msgs:
                        print("message: ", msg.get_data())
                        await msg.ack()
                        headers = msg.get_headers()
                    if error:
                        print(error)
                except Exception as e:
                    print(e)
                    return
        
            try:
                memphis = Memphis()
                await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<broker-token>")
        
                consumer = await memphis.consumer(
                    station_name="<station-name>", consumer_name="<consumer-name>", consumer_group="")
                consumer.consume(msg_handler)
                # Keep your main thread alive so the consumer will keep receiving data
                await asyncio.sleep(5)
        
            except Exception as e:
                print(e)
        
            finally:
                await memphis.close()
        
        if __name__ == '__main__':
            asyncio.run(main())
        `
    }
};
