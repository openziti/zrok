import {Col, Container, Row, Tooltip} from "react-bootstrap";
import {bytesToSize} from "../metrics";
import {Bar, BarChart, CartesianGrid, ResponsiveContainer, XAxis, YAxis} from "recharts";
import moment from "moment/moment";
import React from "react";

const MetricsView = (props) => {
    return (
        <Container>
            <Row>
                <Col>
                    <h3>Last 30 Days:</h3>
                </Col>
            </Row>
            <Row>
                <Col><p>Received: {bytesToSize(props.metrics30.rx)}</p></Col>
                <Col><p>Sent: {bytesToSize(props.metrics30.tx)}</p></Col>
            </Row>
            <Row>
                <Col>
                    <ResponsiveContainer width={"100%"} height={150}>
                        <BarChart data={props.metrics30.data}>
                            <CartesianGrid strokeDasharay={"3 3"} />
                            <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%'}}/>
                            <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%' }}/>
                            <Bar stroke={"#231069"} fill={"#04adef"} dataKey={"rx"} legendType={"circle"}/>
                            <Bar stroke={"#231069"} fill={"#9BF316"} dataKey={"tx"} />
                            <Tooltip />
                        </BarChart>
                    </ResponsiveContainer>
                </Col>
            </Row>
            <Row>
                <Col>
                    <h3>Last 7 Days:</h3>
                </Col>
            </Row>
            <Row>
                <Col><p>Received: {bytesToSize(props.metrics7.rx)}</p></Col>
                <Col><p>Sent: {bytesToSize(props.metrics7.tx)}</p></Col>
            </Row>
            <Row>
                <Col>
                    <ResponsiveContainer width={"100%"} height={150}>
                        <BarChart data={props.metrics7.data}>
                            <CartesianGrid strokeDasharay={"3 3"} />
                            <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%'}}/>
                            <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%' }}/>
                            <Bar stroke={"#231069"} fill={"#04adef"} dataKey={"rx"} legendType={"circle"}/>
                            <Bar stroke={"#231069"} fill={"#9BF316"} dataKey={"tx"} />
                            <Tooltip />
                        </BarChart>
                    </ResponsiveContainer>
                </Col>
            </Row>
            <Row>
                <Col>
                    <h3>Last 24 Hours:</h3>
                </Col>
            </Row>
            <Row>
                <Col><p>Received: {bytesToSize(props.metrics1.rx)}</p></Col>
                <Col><p>Sent: {bytesToSize(props.metrics1.tx)}</p></Col>
            </Row>
            <Row>
                <Col>
                    <ResponsiveContainer width={"100%"} height={150}>
                        <BarChart data={props.metrics1.data}>
                            <CartesianGrid strokeDasharay={"3 3"} />
                            <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%'}}/>
                            <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%' }}/>
                            <Bar stroke={"#231069"} fill={"#04adef"} dataKey={"rx"} legendType={"circle"}/>
                            <Bar stroke={"#231069"} fill={"#9BF316"} dataKey={"tx"} />
                            <Tooltip />
                        </BarChart>
                    </ResponsiveContainer>
                </Col>
            </Row>
        </Container>
    );
}

export default MetricsView;