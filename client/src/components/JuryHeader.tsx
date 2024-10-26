import { useNavigate } from 'react-router-dom';
import { twMerge } from 'tailwind-merge';
import logoutButton from '../assets/logout.svg';

interface JuryHeaderProps {
    /* Whether to show the back to admin button */
    withBack?: boolean;

    /* Whether to show the logout button */
    withLogout?: boolean;

    /* Whether the user is an admin */
    isAdmin?: boolean;
}

const JuryHeader = (props: JuryHeaderProps) => {
    const navigate = useNavigate();

    const backToAdmin = () => navigate('/admin');

    const adminCenter = props.isAdmin ? 'text-center' : '';

    return (
        <div
            className={twMerge(
                'md:px-2 px-4 relative mx-auto pt-6 w-full flex flex-col bg-background',
                props.isAdmin ? 'items-center' : 'md:w-[30rem]'
            )}
        >
            <a
                href="/"
                className={twMerge(
                    'font-bold hover:text-primary duration-200 block max-w-fit',
                    props.isAdmin ? 'text-5xl' : 'text-4xl',
                    adminCenter
                )}
            >
                {props.isAdmin ? 'Jury Admin' : 'Jury'}
            </a>
            <div
                className={twMerge(
                    'font-bold text-primary',
                    props.isAdmin && 'text-[1.5rem]',
                    adminCenter
                )}
            >
                {import.meta.env.VITE_JURY_NAME}
            </div>
            {props.withBack && (
                <div
                    className="absolute top-6 left-6 flex items-center cursor-pointer border-none bg-transparent hover:scale-110 duration-200 text-light text-xl mr-2"
                    onClick={backToAdmin}
                >
                    ◂&nbsp;&nbsp;Back
                </div>
            )}
            {props.withLogout && (
                <div
                    className="absolute top-6 right-6 flex items-center cursor-pointer border-none bg-transparent hover:scale-110 duration-200"
                >
                    <div className="text-light text-xl mr-2"><a href={`${import.meta.env.VITE_API_ORIGIN}/api/auth/keycloak/logout`}>Logout</a></div>
                    <img className="w-4 h-4" src={logoutButton} alt="logout icon" />
                </div>
            )}
        </div>
    );
};

export default JuryHeader;
