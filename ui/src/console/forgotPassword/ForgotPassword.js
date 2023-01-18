import {useState} from "react";
import * as account from '../../api/account';
import {Button, Container, Form, Row} from "react-bootstrap";
import { useLocation } from "react-router-dom";
import SendRequest from "./SendRequest"
import ResetPassword from "./ResetPassword";

const ForgotPassword = (props) => {
    const { search } = useLocation();
    const token = new URLSearchParams(search).get("token")
    console.log(token)
    let forgetPasswordComponent = undefined
    if (token) {
        forgetPasswordComponent = <ResetPassword token={token} />
    } else {
        forgetPasswordComponent = <SendRequest />
    }

    return (
        <div className={"fullscreen"}>{forgetPasswordComponent}</div>
    )
}

export default ForgotPassword;