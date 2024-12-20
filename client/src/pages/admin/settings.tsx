import { useEffect, useState } from 'react';
import { createHeaders, getRequest, postRequest } from '../../api';
import Button from '../../components/Button';
import JuryHeader from '../../components/JuryHeader';
import { errorAlert } from '../../util';
import TextPopup from '../../components/TextPopup';
import Loading from '../../components/Loading';

// Text components
const Section = ({ children: c }: { children: React.ReactNode }) => (
    <h2 className="text-4xl mt-8 mb-2 text-primary">{c}</h2>
);
const SubSection = ({ children: c }: { children: React.ReactNode }) => (
    <h3 className="text-xl font-bold">{c}</h3>
);
const Description = ({ children: c }: { children: React.ReactNode }) => (
    <p className="text-light">{c}</p>
);

const AdminSettings = () => {
    const [clockResetPopup, setClockResetPopup] = useState(false);
    const [dropPopup, setDropPopup] = useState(false);
    const [judgingTimer, setJudgingTimer] = useState('');
    const [minViews, setMinViews] = useState('');
    const [categories, setCategories] = useState('');
    const [batchRankingSize, setBatchRankingSize] = useState('');
    const [loading, setLoading] = useState(true);

    async function getOptions() {
        const res = await getRequest<Options>('/admin/options');
        if (res.status !== 200) {
            errorAlert(res);
            return;
        }
        if (!res.data) {
            alert('error: could not get options data');
            return;
        }

        // Calculate judging timer MM:SS
        const timer = res.data.judging_timer;
        if (timer) {
            const minutes = Math.floor(timer / 60);
            const seconds = timer % 60;
            const timerStr = `${minutes}:${seconds < 10 ? '0' : ''}${seconds}`;
            setJudgingTimer(timerStr);
        }

        // Set categories
        const cats = res.data.categories.join(', ');
        setCategories(cats ?? '');

        // Set min views
        setMinViews(res.data.min_views.toString());

        // Set batch ranking size
        setBatchRankingSize(res.data.batch_ranking_size.toString())
        setLoading(false);
    }

    // Get the previous options
    useEffect(() => {
        getOptions();
    }, []);

    const updateTimer = async () => {
        // Convert judging timer to time
        const [minutes, seconds] = judgingTimer.split(':');
        const timer = judgingTimer === '' ? 0 : parseInt(minutes) * 60 + parseInt(seconds);

        // Check to make sure timer is valid and positive
        if (isNaN(timer)) {
            alert('Invalid timer format!');
            return;
        }
        if (timer < 0) {
            alert('Timer must be a positive number!');
            return;
        }

        // Update the timer
        const res = await postRequest<YesNoResponse>('/admin/timer', {
            judging_timer: timer,
        });
        if (res.status !== 200 || res.data?.yes_no !== 1) {
            errorAlert(res);
            return;
        }

        alert('Timer updated!');
        await getOptions();
    };

    const updateMinViews = async () => {
        // Convert minViews to integer
        const v = parseInt(minViews);
        if (isNaN(v)) {
            alert('Minimum views should be a positive integer!');
            return;
        }

        // Update min views
        const res = await postRequest<YesNoResponse>('/admin/min-views', {
            min_views: v,
        });
        if (res.status !== 200 || res.data?.yes_no !== 1) {
            errorAlert(res);
            return;
        }

        alert('Min views updated!');
        await getOptions();
    };

    const updateCategories = async () => {
        // Split categories by comma and remove empty strings
        const filteredCats = categories
            .split(',')
            .map((cat) => cat.trim())
            .filter((cat) => cat !== '');

        if (filteredCats.length === 0) {
            alert('Categories cannot be empty');
            return
        }
        // Post the new categories
        const res = await postRequest<YesNoResponse>('/admin/categories', {
            categories: filteredCats,
        });
        if (res.status !== 200 || res.data?.yes_no !== 1) {
            errorAlert(res);
            return;
        }

        alert('Categories updated!');
        await getOptions();
    };

    const updateBatchRankingSize = async () => {
        // Convert batchRankingSize to integer
        const r = parseInt(batchRankingSize);
        if (isNaN(r) || r < 2) {
            alert('Minimum batch ranking size should be a positive integer >= 2!');
            return;
        }
        const res = await postRequest<YesNoResponse>('/admin/batch-ranking-size', {
            batch_ranking_size: r,
        });
        if (res.status !== 200 || res.data?.yes_no !== 1) {
            errorAlert(res);
            return;
        }

        alert('Batch Ranking Size updated!');
        await getOptions();
    }

    const resetClock = async () => {
        const res = await postRequest<YesNoResponse>('/admin/clock/reset', null);
        if (res.status !== 200 || res.data?.yes_no !== 1) {
            errorAlert(res);
            return;
        }

        alert('Clock reset!');
        setClockResetPopup(false);
    };

    const dropDatabase = async () => {
        const res = await postRequest<YesNoResponse>('/admin/reset', null);
        if (res.status !== 200 || res.data?.yes_no !== 1) {
            errorAlert(res);
            return;
        }

        alert('Database reset!');
        setDropPopup(false);
    };

    const exportCsv = async (type: string) => {
        const res = await fetch(`${import.meta.env.VITE_API_ORIGIN}/api/admin/export/${type}`, {
            method: 'GET',
            headers: createHeaders(false),
        });

        if (res.status !== 200) {
            const error = 'Error exporting data: ' + res.statusText;
            alert(error);
            console.error(error);
            return;
        }

        saveToFile((await res.blob()) as Blob, type, 'csv');
    };

    const exportByChallenge = async () => {
        const res = await fetch(`${import.meta.env.VITE_API_ORIGIN}/api/admin/export/challenges`, {
            method: 'GET',
            headers: createHeaders(false),
        });

        if (res.status !== 200) {
            const error = 'Error exporting data: ' + res.statusText;
            alert(error);
            console.error(error);
            return;
        }

        saveToFile((await res.blob()) as Blob, 'challenge-projects', 'zip');
    };

    const saveToFile = (blob: Blob, name: string, ext: string) => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.setAttribute('download', `${name}.${ext}`);
        document.body.appendChild(a);
        a.click();
        a.remove();
    };

    return (
        <>
            <JuryHeader withBack withLogout isAdmin />
            <div className="flex flex-col items-start justify-center w-full px-8 py-4 md:px-16 md:py-8">
                <h1 className="text-4xl font-bold">Settings</h1>
                <Section>Judging Settings</Section>

                <SubSection>Reset Main Clock</SubSection>
                <Description>Reset the clock back to 00:00:00</Description>
                <Button
                    type="primary"
                    onClick={() => {
                        setClockResetPopup(true);
                    }}
                    className="mt-4 w-auto md:w-auto px-4 py-2 mb-8"
                >
                    Reset
                </Button>

                <SubSection>Set Judging Timer</SubSection>
                <Description>
                    Set how long judges have to view each project. This will reflect on the timer
                    that shows on the judging page. Leave this field blank (or 0) if you do not wish
                    to have a timer for each judge.
                </Description>
                <input
                    className="w-full h-14 px-4 text-2xl border-lightest border-2 rounded-sm focus:border-primary focus:border-4 focus:outline-none"
                    type="string"
                    placeholder="MM:SS"
                    value={judgingTimer}
                    onChange={(e) => {
                        setJudgingTimer(e.target.value);
                    }}
                />
                <Button
                    type="primary"
                    onClick={updateTimer}
                    className="mt-4 w-auto md:w-auto px-4 py-2 mb-8"
                >
                    Update Timer
                </Button>

                <SubSection>Set Categories</SubSection>
                <Description>
                    Set the categories that the judges will be scoring each project on. Please
                    separate each category with a comma.
                </Description>
                <input
                    className="w-full h-14 px-4 text-xl border-lightest border-2 rounded-sm focus:border-primary focus:border-4 focus:outline-none"
                    type="string"
                    placeholder="Cat 1, Cat 2, Cat 3, ..."
                    value={categories}
                    onChange={(e) => {
                        setCategories(e.target.value);
                    }}
                />
                <Button
                    type="primary"
                    onClick={updateCategories}
                    className="mt-4 w-auto md:w-auto px-4 py-2 mb-8"
                >
                    Update Categories
                </Button>

                <SubSection>Set Batch Ranking Size (BRS)</SubSection>
                <Description>
                    Set how many projects judges rank at a time (must be at least 2 obviously).
                    Judges can rank and reorder projects freely before submitting a batch of the specified size.
                    Avoid changing this while judging is ongoing as judges may have already accumulated more/less projects.
                </Description>
                <input
                    className="w-full h-14 px-4 text-2xl border-lightest border-2 rounded-sm focus:border-primary focus:border-4 focus:outline-none"
                    type="number"
                    min="2"
                    placeholder="8"
                    value={batchRankingSize}
                    onChange={(e) => {
                        setBatchRankingSize(e.target.value.toString());
                    }}
                />
                <Button
                    type="primary"
                    onClick={updateBatchRankingSize}
                    className="mt-4 w-auto md:w-auto px-4 py-2"
                >
                    Update Batch Ranking Size
                </Button>

                <Section>Judging Parameters</Section>

                <SubSection>Set Minimum Project Views</SubSection>
                <Description>
                    Set the minimum amount of times that a project should be seen during judging.
                    This will ensure all projects get seen at LEAST this many times before switching
                    over to the optimal method of assigning projects. Set to 0 to ignore this
                    condition (recommended: 3-5).
                </Description>
                <input
                    className="w-full h-14 px-4 text-2xl border-lightest border-2 rounded-sm focus:border-primary focus:border-4 focus:outline-none"
                    type="string"
                    placeholder="Enter integer..."
                    value={minViews}
                    onChange={(e) => {
                        setMinViews(e.target.value);
                    }}
                />
                <Button
                    type="primary"
                    onClick={updateMinViews}
                    className="mt-4 w-auto md:w-auto px-4 py-2 mb-8"
                >
                    Update Min Views
                </Button>

                <Section>Export Data</Section>

                <SubSection>Export Collection</SubSection>
                <Description>Export each collection individually as a CSV download.</Description>
                <div className="flex">
                    <Button
                        type="primary"
                        onClick={() => {
                            exportCsv('judges');
                        }}
                        className="mt-4 w-auto md:w-auto px-4 py-2 mr-4"
                    >
                        Export Judges
                    </Button>
                    <Button
                        type="primary"
                        onClick={() => {
                            exportCsv('projects');
                        }}
                        className="mt-4 w-auto md:w-auto px-4 py-2 mr-4"
                    >
                        Export Projects
                    </Button>
                    <Button
                        type="primary"
                        onClick={exportByChallenge}
                        className="mt-4 w-auto md:w-auto px-4 py-2 mr-4"
                    >
                        Export by Challenges
                    </Button>
                    <Button
                        type="primary"
                        onClick={() => {
                            exportCsv('rankings');
                        }}
                        className="mt-4 w-auto md:w-auto px-4 py-2"
                    >
                        Export Rankings
                    </Button>
                </div>

                <Section>Reset Database</Section>

                <SubSection>THIS WILL DELETE THE ENTIRE DATABASE</SubSection>
                <Description>
                    Mostly used for testing purposes/before the event if you want to reset
                    everything bc something got messed up. Do NOT use this during judging (duh).
                </Description>
                <Button
                    type="error"
                    onClick={() => {
                        setDropPopup(true);
                    }}
                    className="mt-4 w-auto md:w-auto px-4 py-2"
                >
                    Drop Database
                </Button>
            </div>
            <TextPopup
                enabled={clockResetPopup}
                setEnabled={setClockResetPopup}
                onSubmit={resetClock}
                submitText="Reset"
                title="Heads Up!"
                red
            >
                Are you sure you want to reset the main clock? This will reset the clock to 00:00:00
            </TextPopup>
            <TextPopup
                enabled={dropPopup}
                setEnabled={setDropPopup}
                onSubmit={dropDatabase}
                submitText="RESET DATA"
                title="Heads Up!"
                red
            >
                THIS WILL ACTUALLY DELETE ALL DATA!!!!! YOU NEED TO BE ABSOLUTELY SURE YOU WANT TO
                DO THIS. THIS IS YOUR LAST WARNING!
            </TextPopup>
            <Loading disabled={!loading} />
        </>
    );
};

export default AdminSettings;
