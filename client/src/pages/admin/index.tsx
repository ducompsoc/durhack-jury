import { useEffect, useState } from 'react';
import AdminStatsPanel from '../../components/admin/AdminStatsPanel';
import AdminTable from '../../components/admin/tables/AdminTable';
import AdminToggleSwitch from '../../components/admin/AdminToggleSwitch';
import AdminToolbar from '../../components/admin/AdminToolbar';
import JuryHeader from '../../components/JuryHeader';
import Loading from '../../components/Loading';
import {getRequest, postRequest} from '../../api';
import { errorAlert } from '../../util';
import { useNavigate } from 'react-router-dom';
import Button from '../../components/Button';

// TODO: Add FAB to 'return to top'
// TODO: Make pause button/settings have hover effects
const Admin = () => {
    const navigate = useNavigate();
    const [showProjects, setShowProjects] = useState(true);
    const [loading, setLoading] = useState(true);
    const [judgingEnded, setJudgingEnded] = useState(false);
    const [numJudges, setNumJudges] = useState(0);

    useEffect(() => {
        // Check if user logged in
        async function checkLoggedIn() {
            const loggedInRes = await postRequest<OkResponse>('/admin/auth', null);
            if (loggedInRes.status === 401) {
                console.error(`Admin is not logged in!`);
                navigate('/');
                return;
            }
            if (loggedInRes.status === 200) {
                setLoading(false);
                return;
            }

            errorAlert(loggedInRes);
        }
        checkLoggedIn();

        async function getNumJudges() {
            const judgeListRes = await getRequest<Judge[]>('/judge/list')
            if (judgeListRes.status !== 200) {
                errorAlert(judgeListRes);
                return;
            }
            setNumJudges(judgeListRes.data?.length as number);

        }
        getNumJudges();

        async function checkJudgingEnded() {
            // const judgingEndedRes = await getRequest<OkResponse>('/admin/end_judging')
            setJudgingEnded(false)
        }
        checkJudgingEnded();
    }, []);

    function endJudging() {
        let confirmed = window.confirm("Are you sure you want to end judging? This cannot be undone.\n" +
            "Judges will not be able to request new projects to rank and must submit rankings.")
        if (confirmed) {
            endJudgingReq().then(success => {
                if (success) {
                    alert("Judging has now been ended. Judges will be notified and made to submit their rankings. " +
                        "Wait until all have submitted before recording final results.")
                    setJudgingEnded(true)
                } else {
                    alert("Failed to end judging.")
                }
            })
        }
    }

    async function endJudgingReq() {
        console.log("Requesting server to end judging.")
        // const endJudgingRes = await postRequest<OkResponse>('/admin/end_judging', null)
        // lucatodo: pause and remove clock
        return true
    }

    if (loading) {
        return <Loading disabled={!loading} />;
    }
    return (
        <>
            <JuryHeader withLogout isAdmin />
            <Button
                type="outline"
                onClick={() => {
                    navigate('/admin/settings');
                }}
                className="absolute top-6 left-[16rem] w-40 md:w-52 text-lg py-2 px-1 hover:scale-100 focus:scale-100 rounded-md font-bold"
            >Settings</Button>
            <AdminStatsPanel />
            <div className="w-full flex flex-row center items-center justify-center my-5">
                <Button
                    type="error"
                    onClick={endJudging}
                    disabled={judgingEnded}
                    bold
                    className="w-1/3 md:w-1/4"
                >{judgingEnded ? `Submitted judges: ${1}/${numJudges}` : "End Judging"}</Button>
            </div>
            <AdminToggleSwitch state={showProjects} setState={setShowProjects} />
            <AdminToolbar showProjects={showProjects} />
            <AdminTable showProjects={showProjects} />
        </>
    );
};

export default Admin;
