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
    const [submittedJudges, setSubmittedJudges] = useState(0);

    useEffect(() => {
        // Check if user logged in
        async function checkLoggedIn() {
            const loggedInRes = await postRequest<YesNoResponse>('/admin/auth', null);
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
        checkSubmittedJudges();

        async function checkJudgingEnded() {
            const judgingEndedRes = await getRequest<JudgingEnded>('/admin/end-judging')
            if (judgingEndedRes.status !== 200) {
                errorAlert(judgingEndedRes);
                return;
            }
            setJudgingEnded(judgingEndedRes.data?.judging_ended as boolean);
        }
        checkJudgingEnded();
    }, []);

    function endJudging() {
        let confirmed = window.confirm("Are you sure you want to end judging? This cannot be undone.\n" +
            "Judges will not be able to request new projects to rank and must submit rankings.");
        if (confirmed) {
            endJudgingReq().then(success => {
                if (success) {
                    alert("Judging has now been ended. Judges will be notified and made to submit their rankings. " +
                        "Wait until all have submitted before recording final results.");
                    setJudgingEnded(true);
                    checkSubmittedJudges();
                } else {
                    alert("Failed to end judging.");
                    setJudgingEnded(false);
                }
            })
        }
    }

    async function endJudgingReq() {
        console.log("Requesting server to end judging.")
        const endJudgingRes = await postRequest<YesNoResponse>('/admin/end-judging', null)
        return endJudgingRes.status === 200;
    }

    async function checkSubmittedJudges() {
        console.log("Refreshing submitted judge count by counting current_projects array lengths")
        const justListRes = await getRequest<Judge[]>('/judge/list')
        if (justListRes.status !== 200) {
            errorAlert(justListRes);
            return;
        }
        if (justListRes.data){
            let numSubmitted = 0
            justListRes.data.forEach(j => {
                if (j.current_rankings.length == 0) numSubmitted++
            })
            setSubmittedJudges(numSubmitted)
        }
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
            <div className="w-full grid grid-cols-3 justify-center justify-items-center items-center my-5">
                <div></div>
                <Button
                    type="error"
                    onClick={endJudging}
                    disabled={judgingEnded}
                    bold
                    className="justify-self-stretch md:w-full w-full"
                >{judgingEnded ? `Submitted judges: ${submittedJudges}/${numJudges}` : "End Judging"}</Button>
                <div hidden={!judgingEnded} onClick={checkSubmittedJudges} className="justify-self-start cursor-pointer" title="Refresh submitted judges">🔁</div>
            </div>
            <AdminToggleSwitch state={showProjects} setState={setShowProjects} />
            <AdminToolbar showProjects={showProjects} />
            <AdminTable showProjects={showProjects} />
        </>
    );
};

export default Admin;
