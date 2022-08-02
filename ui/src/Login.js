import {useState} from "react";

const Login = (props) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const handleSubmit = async e => {
        e.preventDefault()
        props.loginSuccess({
            email: email,
        })
    };

    return (
        <form onSubmit={handleSubmit}>
            <label htmlFor="email">email: </label>
            <input type="text" value={email} placeholder="enter an email" onChange={({ target }) => setEmail(target.value)}/>
            <div>
                <label htmlFor="password">password: </label>
                <input type="password" value={password} placeholder="enter a password" onChange={({ target}) => setPassword(target.value)}/>
            </div>
            <button type="submit">Log In</button>
        </form>
    );
}

export default Login;