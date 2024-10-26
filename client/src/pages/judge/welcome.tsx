import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Container from '../../components/Container';
import JuryHeader from '../../components/JuryHeader';
import Checkbox from '../../components/Checkbox';
import Button from '../../components/Button';
import { getRequest, postRequest } from '../../api';
import { errorAlert } from '../../util';
import Loading from '../../components/Loading';

const JudgeWelcome = () => {
    const navigate = useNavigate();
    const [judge, setJudge] = useState<Judge | null>(null);
    const [checkRead, setCheckRead] = useState(false);
    const [checkEmail, setCheckEmail] = useState(false);

    // Verify user is logged in and read welcome before proceeding
    useEffect(() => {
        async function fetchData() {
            // Check to see if the user is logged in
            const loggedInRes = await postRequest<YesNoResponse>('/judge/auth', null);
            if (loggedInRes.status !== 200) {
                errorAlert(loggedInRes);
                return;
            }
            if (loggedInRes.data?.yes_no !== 1) {
                console.error(`Judge is not logged in!`);
                navigate('/');
                return;
            }

            // Get the name & email of the user from the server
            const judgeRes = await getRequest<Judge>('/judge');
            if (judgeRes.status !== 200) {
                errorAlert(judgeRes);
                return;
            }
            setJudge(judgeRes.data as Judge);
        }

        fetchData();
    }, []);

    // Read the welcome message and mark that the user has read it
    const readWelcome = async () => {
        if (!checkRead || !checkEmail) {
            alert(
                'Please read the welcome message and confirm by checking the boxes below before proceeding.'
            );
            return;
        }

        // POST to server to mark that the user has read the welcome message
        const readWelcomeRes = await postRequest<YesNoResponse>('/judge/welcome', null);
        if (readWelcomeRes.status !== 200) {
            errorAlert(readWelcomeRes);
            return;
        }

        navigate('/judge');
    };

    if (!judge) return <Loading disabled={judge !== null} />;

    return (
        <>
            <JuryHeader withLogout />
            <Container noCenter className="mx-7">
                <h1 className="text-2xl my-2">Hello, {judge.name}!</h1>
                <h2 className="text-lg font-bold">PLEASE READ THE FOLLOWING:</h2>
                <p className="my-2">
                    Welcome to Jury, an innovative judging system that uses a dynamic ranking system
                    to facilitate hackathon judging.
                </p>
                <p className="my-2">
                    Originally created by&nbsp;
                    <a className="text-primary" href="https://github.com/hackutd/jury" target="_blank">MichaelZhao21</a>
                    &nbsp;inspired by&nbsp;
                    <a className="text-primary" href="https://github.com/anishathalye/gavel" target="_blank">Gavel by
                        anishathalye</a>
                    &nbsp;and adapted for use in DurHack 2024 by&nbsp;
                    <a className="text-primary" href="https://github.com/ducompsoc/durhack-jury" target="_blank">Luca
                        (tameTNT)</a>.
                </p>
                <p className="my-2">
                    Once you get started, you will be presented with a project and its location.
                    Please go to that project and listen to their presentation
                    (use the inbuilt timer to time their pitch if you like).
                    You can also make notes in the textbox at the bottom of the screen.
                    Once completed, please score their project on the respective categories and click "Done".
                    Your notes, as well as your scores, are saved and can be reviewed and changed any time
                    by clicking on the project name.
                </p>
                <p className="my-2">
                    Once you have scored a project, you will be taken to the ranking screen. Here,
                    you can rank the project relative to others you have seen. You can also view the
                    projects you&apos;ve seen previously and adjust their scores
                    (rankings are saved so rank as you go!).
                </p>
                <p className="my-2">
                    You will rank projects in
                    batches (size determined by the organizers). Once you have seen the set number of projects,
                    you will not be able to see any more and must rank those that you have seen so far.
                    Press submit to submit your current rankings as a batch and be able to see more projects.
                    When judging is ended by an organiser, you will not be able to view any further projects but will
                    be able to rank and submit any you have already seen, even if this is less than the set batch size.
                </p>
                <p className="my-2">
                    If a team is busy being judged, click the &apos;busy&apos; button. This will NOT
                    impact their rating and you may be presented with this team again.
                </p>
                <p className="my-2">
                    If a team is absent or you suspect a team may be cheating, please report it to
                    the organizers with the &apos;flag&apos; button. We will look into the matter
                    and take the proper action.
                </p>
                <p className="my-2">
                    If you encounter any issues with the system, please contact a member of the
                    organising team.
                </p>
                <Checkbox checked={checkRead} onChange={setCheckRead}>
                    Before you continue, please acknowledge that you have read and understand the
                    above instructions.
                </Checkbox>
                <Checkbox checked={checkEmail} onChange={setCheckEmail}>
                    I certify that my email is <span className="text-primary">[{judge.email}]</span>
                    . If this is not your email, contact an organizer immediately.
                </Checkbox>
                <div className="flex justify-center py-4">
                    <Button
                        type="primary"
                        disabled={!checkRead || !checkEmail}
                        onClick={readWelcome}
                        className="my-2"
                    >
                        Continue
                    </Button>
                </div>
            </Container>
        </>
    );
};

export default JudgeWelcome;
