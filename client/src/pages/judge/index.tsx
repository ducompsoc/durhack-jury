import { useEffect, useState } from 'react';
import Container from '../../components/Container';
import JuryHeader from '../../components/JuryHeader';
import { useNavigate } from 'react-router-dom';
import Button from '../../components/Button';
import StatBlock from '../../components/StatBlock';
import Loading from '../../components/Loading';
import { getRequest, postRequest } from '../../api';
import { errorAlert } from '../../util';
import {
    DndContext,
    DragEndEvent,
    DragOverEvent,
    DragOverlay,
    DragStartEvent,
    KeyboardSensor,
    UniqueIdentifier,
    closestCenter,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import { arrayMove, sortableKeyboardCoordinates } from '@dnd-kit/sortable';
import Droppable from '../../components/judge/dnd/Droppable';
import RankItem from '../../components/judge/dnd/RankItem';
import CustomPointerSensor from '../../components/judge/dnd/CustomPointerSensor';

const Judge = () => {
    const navigate = useNavigate();
    const [judge, setJudge] = useState<Judge | null>(null);
    const [ranked, setRanked] = useState<SortableJudgedProject[]>([]);
    const [unranked, setUnranked] = useState<SortableJudgedProject[]>([]);
    const [allRanked, setAllRanked] = useState(false)
    const [batchRankingSize, setBatchRankingSize] = useState(0);
    const [judgingIsOver, setJudgingIsOver] = useState(false);
    const [nextButtonDisabled, setNextButtonDisabled] = useState(false);
    const [nextButtonHelperText, setNextButtonHelperText] = useState('');
    const [loaded, setLoaded] = useState(false);
    const [projCount, setProjCount] = useState(0);
    const [activeId, setActiveId] = useState<number | null>(null);
    const [activeDropzone, setActiveDropzone] = useState<string | null>(null);
    const sensors = useSensors(
        useSensor(CustomPointerSensor, {
            activationConstraint: {
                distance: 5,
            },
        }),
        useSensor(KeyboardSensor, {
            coordinateGetter: sortableKeyboardCoordinates,
        })
    );

    // Verify user is logged in and read welcome before proceeding
    useEffect(() => {
        async function fetchData() {
            // Check to see if the user is logged in
            const loggedInRes = await postRequest<YesNoResponse>('/judge/auth', null);
            if (loggedInRes.status === 401) {
                console.error(`Judge is not logged in!`);
                navigate('/');
                return;
            }
            if (loggedInRes.status !== 200) {
                errorAlert(loggedInRes);
                return;
            }
            if (loggedInRes.data?.yes_no !== 1) {
                console.error(`Judge is not logged in!`);
                navigate('/');
                return;
            }

            // Check for read welcome
            const readWelcomeRes = await getRequest<YesNoResponse>('/judge/welcome');
            if (readWelcomeRes.status !== 200) {
                errorAlert(readWelcomeRes);
                return;
            }
            const readWelcome = readWelcomeRes.data?.yes_no === 1;
            if (!readWelcome) {
                navigate('/judge/welcome');
            }

            // Get the name & email of the user from the server
            const judgeRes = await getRequest<Judge>('/judge');
            if (judgeRes.status !== 200) {
                errorAlert(judgeRes);
                return;
            }
            const judge: Judge = judgeRes.data as Judge;
            setJudge(judge);

            // Get the project count
            const projCountRes = await getRequest<ProjectCount>('/project/count');
            if (projCountRes.status !== 200) {
                errorAlert(projCountRes);
                return;
            }
            setProjCount(projCountRes.data?.count as number);

            // Get Batch Ranking Size
            const batchRankingSizeRes = await getRequest<BatchRankingSize>('/brs');
            if (batchRankingSizeRes.status !== 200) {
                errorAlert(batchRankingSizeRes);
                return;
            }
            setBatchRankingSize(batchRankingSizeRes.data?.brs as number);

            const judgingOverRes = await getRequest<YesNoResponse>('/check-judging-over')
            if (judgingOverRes.status !== 200) {
                errorAlert(judgingOverRes);
                return;
            }
            let judgingOverResBool = Boolean(judgingOverRes.data?.yes_no)
            setJudgingIsOver(judgingOverResBool);
        }

        fetchData();
    }, []);

    // Load all projects when judge loads
    useEffect(() => {
        if (!judge) return;

        const currentBatchProjects = judge.seen_projects.map((p, i) => ({
            id: i + 1,
            ...p,
        })).filter((p) => // filter out projects that were ranked in a previous batch
            judge.past_rankings.flat().every((r) => r !== p.project_id)
        );

        const rankedProjects = judge.current_rankings.map((r) =>
            currentBatchProjects.find((p) => p.project_id === r)
        ) as SortableJudgedProject[];
        const unrankedProjects = currentBatchProjects.filter((p) =>
            judge.current_rankings.every((r) => r !== p.project_id)
        );
        unrankedProjects.reverse();

        setRanked(rankedProjects);
        setUnranked(unrankedProjects);

        setLoaded(true);
    }, [judge]);

    // Trigger button state ranking batch logic updates when `batchRankingSize` is set (>0) and/or whenever `ranked` or `unranked` states chance
    useEffect(() => {
        if (batchRankingSize > 0) {
            setAllRanked((ranked.length === batchRankingSize || judgingIsOver) && unranked.length === 0);

            if (ranked.length + unranked.length === batchRankingSize) {
                setNextButtonHelperText('Rank and submit your current batch to move on');
                setNextButtonDisabled(true);
            } else {
                setNextButtonHelperText('');
                if (!judgingIsOver) setNextButtonDisabled(false);
            }
        }
        if (judgingIsOver) {
            setNextButtonDisabled(true);
        }
    }, [batchRankingSize, judgingIsOver, ranked, unranked, loaded, judgingIsOver]);

    if (!loaded) return <Loading disabled={!loaded} />;

    // Lets the user take a break
    const takeBreak = async () => {
        // Check if the user is allowed to take a break
        if (judge?.current == null) {
            alert('You are already taking a break!');
            return;
        }

        const res = await postRequest<YesNoResponse>('/judge/break', null);
        if (res.status !== 200) {
            errorAlert(res);
            return;
        }

        alert('You can now take a break! Press "Next project" to continue judging.');
    };

    const handleDragStart = (event: DragStartEvent) => {
        const { active } = event;
        setActiveId(active.id as number);
    };

    const handleDragOver = (event: DragOverEvent) => {
        const { active, over } = event;
        const { id } = active;

        if (over === null) {
            setActiveId(null);
            return;
        }
        const { id: overId } = over;

        const activeRanked = isRankedObject(id);
        const overRanked = isRankedObject(overId);

        setActiveDropzone(overRanked ? 'ranked' : 'unranked');

        // If moving to new container, swap the item to the new list
        if (activeRanked !== overRanked) {
            const activeContainer = activeRanked ? ranked : unranked;
            const overContainer = overRanked ? ranked : unranked;
            const oldIndex = activeContainer.findIndex((i) => i.id === active.id);
            const newIndex = overContainer.findIndex((i) => i.id === over.id);
            const proj = activeContainer[oldIndex];
            // @ts-ignore
            const newActive = activeContainer.toSpliced(oldIndex, 1);
            // @ts-ignore
            const newOver = overContainer.toSpliced(newIndex, 0, proj);
            if (activeRanked) {
                setRanked(newActive);
                setUnranked(newOver);
            } else {
                setRanked(newOver);
                setUnranked(newActive);
            }
        }
    };

    const handleDragEnd = (event: DragEndEvent) => {
        const { active, over } = event;
        const { id } = active;

        if (over === null) {
            setActiveId(null);
            return;
        }
        const { id: overId } = over;

        const activeRanked = isRankedObject(id);
        const overRanked = isRankedObject(overId);

        if (activeRanked === overRanked) {
            const currProjs = activeRanked ? ranked : unranked;

            const oldIndex = currProjs.findIndex((i) => i.id === active.id);
            const newIndex = currProjs.findIndex((i) => i.id === over.id);
            const newProjects: SortableJudgedProject[] = arrayMove(currProjs, oldIndex, newIndex);
            activeRanked ? setRanked(newProjects) : setUnranked(newProjects);

            if (activeRanked) saveSort(newProjects);
            else saveSort(ranked);
        } else {
            saveSort(ranked);
        }

        setActiveDropzone(null);
        setActiveId(null);
    };

    // dnd-kit is strange. For active/over ids, it is a number most of the time,
    // representing the ID of the item that we are hovering over.
    // However, if the user is hovering NOT on an item, it will set the ID
    // to the ID of the droppable container ?!??!
    // Strange indeed.
    function isRankedObject(id: UniqueIdentifier) {
        // If drop onto the zone (id would be string)
        if (isNaN(Number(id))) {
            return id === 'ranked';
        }

        // Otherwise if dropped onto a specific object
        const ro = ranked.find((a) => a.id === id);
        return !!ro;
    }

    const saveSort = async (projects: SortableJudgedProject[]) => {
        // Save the rankings
        const saveRes = await postRequest<YesNoResponse>('/judge/rank', {
            ranking: projects.map((p) => p.project_id),
        });
        if (saveRes.status !== 200) {
            errorAlert(saveRes);
            return;
        }
    };

    const submitBatch = async () => {
        if (ranked.length !== batchRankingSize && !judgingIsOver) {
            alert(`You can only submit rankings in batches of ${batchRankingSize} projects.`)
            return
        }
        if (ranked.length === 0) {
            alert('You cannot submit an empty batch.')
            return
        }

        const submitRes = await postRequest<YesNoResponse>('/judge/submit-batch-ranking', {
            batch_ranking: ranked.map((p) => p.project_id),
        });
        if (submitRes.status !== 200) {
            errorAlert(submitRes);
            return;
        } else if (submitRes.status === 200) {
            alert('Ranking batch submitted successfully!')
            window.location.reload()
        }
    }

    return (
        <>
            <JuryHeader withLogout />
            <Container noCenter className="px-2 pb-4">
                <div className="w-full text-lg text-center italic bg-error" hidden={!judgingIsOver}>
                    <p>Judging has been ended. You can no longer view new projects.</p>
                    <p>Please rank your previously seen projects and submit.</p>
                </div>
                <h1 className="text-2xl my-2">Welcome, {judge?.name}!</h1>
                <div className="w-full mb-6">
                    <Button type="primary" full square href="/judge/live" disabled={nextButtonDisabled}>
                        Next Project
                        <p className="text-sm italic">{nextButtonHelperText}</p>
                    </Button>
                    <div className="flex align-center justify-center mt-4">
                        <Button type="outline" square onClick={takeBreak} disabled={judgingIsOver} className="text-lg p-2">
                            I want to take a break!
                        </Button>
                    </div>
                </div>
                <div className="flex justify-evenly">
                    <StatBlock name="Seen" value={judge?.seen_projects.length as number}/>
                    <StatBlock name="Submitted Batches" value={judge?.past_rankings.length as number}/>
                    <StatBlock name="Total Projects" value={projCount}/>
                </div>
                <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragStart={handleDragStart}
                    onDragOver={handleDragOver}
                    onDragEnd={handleDragEnd}
                >
                    <h2 className="text-primary text-xl font-bold mt-4">Ranked Projects</h2>
                    <p className="text-light text-sm">
                        Click on titles to edit scores and see details.
                    </p>
                    <p className="text-light text-sm italic">
                        NB: Your rankings are not counted until they are submitted as a batch. Reload if dragging breaks.
                    </p>
                    <div className="h-[1px] w-full bg-light my-2"></div>
                    <Droppable id="ranked" projects={ranked} active={activeDropzone} />

                    <h2 className="text-primary text-xl font-bold mt-4">Unranked Projects</h2>
                    <p className="text-light text-sm">
                        Projects will be sorted in reverse chronological order.
                    </p>
                    <div className="h-[1px] w-full bg-light my-2"></div>
                    <Droppable id="unranked" projects={unranked} active={activeDropzone} />

                    <DragOverlay>
                        {activeId ? (
                            <RankItem
                                item={
                                    unranked.find((p) => p.id === activeId) ??
                                    (ranked.find((p) => p.id === activeId) as SortableJudgedProject)
                                }
                                ranking={ranked.findIndex((p) => p.id === activeId) + 1}
                            />
                        ) : null}
                    </DragOverlay>
                </DndContext>
                <div className="w-full mt-4">
                    <div className="justify-center text-light text-sm italic text-center">
                        Please rank all your projects to submit.<br/>
                        <p hidden={judgingIsOver}>You can only submit rankings in batches of {batchRankingSize} projects.</p>
                    </div>
                    <Button type="primary" full square className="mt-1" disabled={!allRanked} onClick={submitBatch}>
                        Submit Rankings
                        <p className="text-sm italic" hidden={judgingIsOver}>and move onto next batch</p>
                        <p className="text-sm italic" hidden={!judgingIsOver}>and finish judging. Thank you for your hard work!</p>
                    </Button>
                </div>
            </Container>
        </>
    );
};

export default Judge;
