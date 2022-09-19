import { useParams } from 'react-router-dom';

const Register = () => {
    const { token } = useParams();

    return (
        <div>
            <h1>Register!</h1>
            <p>token = "{token}"</p>
        </div>
    )
}

export default Register;