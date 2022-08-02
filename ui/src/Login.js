import {useState} from "react";
import * as identity from './api/identity';

const Login = (props) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const handleSubmit = async e => {
        e.preventDefault()
        identity.login({body: {"email": email, "password": password}})
            .then(resp => {
                if(!resp.error) {
                    props.loginSuccess({
                        email: email,
                        token: resp.token
                    })
                    console.log('login succeeded', resp)
                } else {
                    console.log('login failed')
                }
            })
            .catch((resp) => {
                console.log('login failed', resp)
            });
    };

    return (
        <form onSubmit={handleSubmit}>
            <label htmlFor="email">email: </label>
            <input type="text" value={email} placeholder="enter an email" onChange={({ target }) => setEmail(target.value)}/>
            <div>
                <label htmlFor="password">password: </label>
                <input type="password" value={password} placeholder="enter a password" onChange={({ target }) => setPassword(target.value)}/>
            </div>
            <button type="submit">Log In</button>
        </form>
    );
}

export default Login;