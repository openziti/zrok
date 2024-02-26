import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import {ApiClient, AccountApi, MetadataApi} from "./api/src"
import App from "./App";

export const zrokClient = new ApiClient()
export const accountApi = new AccountApi(zrokClient)
export const metadataApi = new MetadataApi(zrokClient)
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

export function updateApiKey() {
    getApiKey().then(key => {
        // Configure API key authorization: key

        let v = zrokClient.authentications['key'];
        v.apiKey = key.apiKey
    })
}

updateApiKey();

