import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import {ApiClient} from "./api/src"
import App from "./App";

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
    <App />
);

function getApiKey() {
    const localUser = JSON.parse(localStorage.getItem("user"))
    if(localUser) {
        return Promise.resolve({ apiKey: localUser.token });
    } else {
        throw new Error("token not available");
    }
}

getApiKey().then(key => {
    let defaultClient = ApiClient.instance;
    // Configure API key authorization: key
    let k = defaultClient.authentications['key'];
    k.apiKey = key.apiKey;
    
})