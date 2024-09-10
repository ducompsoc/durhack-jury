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
    DragOverlay,
    DragStartEvent,
    KeyboardSensor,
    PointerSensor,
    closestCenter,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import {
    SortableContext,
    arrayMove,
    sortableKeyboardCoordinates,
    verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import SortableItem from '../../components/judge/SortableItem';

const Judge = () => {
    const navigate = useNavigate();
    const [judge, setJudge] = useState<Judge | null>(null);
    const [projects, setProjects] = useState<SortableJudgedProject[]>([]);
    const [allRanked, setAllRanked] = useState(false)
    const [loaded, setLoaded] = useState(false);
    const [projCount, setProjCount] = useState(0);
    const [activeId, setActiveId] = useState<number | null>(null);
    const sensors = useSensors(
        useSensor(PointerSensor, {
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
            const loggedInRes = await postRequest<OkResponse>('/judge/auth', 'judge', null);
            if (loggedInRes.status === 401) {
                console.error(`Judge is not logged in!`);
                navigate('/judge/login');
                return;
            }
            if (loggedInRes.status !== 200) {
                errorAlert(loggedInRes);
                return;
            }
            if (loggedInRes.data?.ok !== 1) {
                console.error(`Judge is not logged in!`);
                navigate('/judge/login');
                return;
            }

            // Check for read welcome
            const readWelcomeRes = await getRequest<OkResponse>('/judge/welcome', 'judge');
            if (readWelcomeRes.status !== 200) {
                errorAlert(readWelcomeRes);
                return;
            }
            const readWelcome = readWelcomeRes.data?.ok === 1;
            if (!readWelcome) {
                navigate('/judge/welcome');
            }

            // Get the name & email of the user from the server
            const judgeRes = await getRequest<Judge>('/judge', 'judge');
            if (judgeRes.status !== 200) {
                errorAlert(judgeRes);
                return;
            }
            const judge: Judge = judgeRes.data as Judge;
            setJudge(judge);

            // Get the project count
            const projCountRes = await getRequest<ProjectCount>('/project/count', 'judge');
            if (projCountRes.status !== 200) {
                errorAlert(projCountRes);
                return;
            }
            setProjCount(projCountRes.data?.count as number);
        }

        fetchData();
    }, []);

    useEffect(() => {
        if (!judge) return;

        const allProjects = judge.seen_projects.map((p, i) => ({
            id: i + 1,
            ...p,
        }));

        const rankedProjects = judge.rankings.map((r) =>
            allProjects.find((p) => p.project_id === r)
        ) as SortableJudgedProject[];
        const unrankedProjects = allProjects.filter((p) =>
            judge.rankings.every((r) => r !== p.project_id)
        );

        // Create dummy project
        const dummy = {
            id: -1,
            project_id: '',
            categories: {},
            notes: '',
            name: 'Unsorted Projects',
            location: 0,
            description: '',
        };

        const combinedProjects = [...rankedProjects, dummy, ...unrankedProjects];

        setProjects(combinedProjects);
        setAllRanked(rankedProjects.length === 3);  // lucatodo: use rank batch size variable;
        // lucatodo: rankedProjects won't be variable we want to find length of. new variable probable: batchProjects?
        setLoaded(true);
    }, [judge]);

    if (!loaded) return <Loading disabled={!loaded} />;

    const takeBreak = async () => {
        // Check if the user is allowed to take a break
        if (judge?.current == null) {
            alert('You are already taking a break!');
            return;
        }

        const res = await postRequest<OkResponse>('/judge/break', 'judge', null);
        if (res.status !== 200) {
            errorAlert(res);
            return;
        }

        alert('You can now take a break! Press "Next project" to continue judging.');
    };

    const handleDragStart = (event: DragStartEvent) => {
        const { active } = event;
        console.log(active.id);
        setActiveId(active.id as number);
    };

    const handleDragEnd = (event: DragEndEvent) => {
        const { active, over } = event;

        if (over != null) {
            const oldIndex = projects.findIndex((i) => i.id === active.id);
            const newIndex = projects.findIndex((i) => i.id === over.id);
            const newProjects = arrayMove(projects, oldIndex, newIndex);
            setProjects(newProjects);
            saveSort(newProjects);
        }

        setActiveId(null);
    };

    const saveSort = async (newProjects: SortableJudgedProject[]) => {
        // Split index
        const splitIndex = newProjects.findIndex((p) => p.id === -1);
        setAllRanked(newProjects.length>0 && splitIndex === newProjects.length-1);  // lucatodo: use rank batch size variable equality, not >0

        // Get the ranked projects
        const rankedProjects = newProjects.slice(0, splitIndex);

        // Save the rankings
        const saveRes = await postRequest<OkResponse>('/judge/rank', 'judge', {
            ranking: rankedProjects.map((p) => p.project_id),
        });
        if (saveRes.status !== 200) {
            errorAlert(saveRes);
            return;
        }
    };

    return (
        <>
            <JuryHeader withLogout />
            <Container noCenter className="px-2 pb-4">
                <h1 className="text-2xl my-2">Welcome, {judge?.name}!</h1>
                <div className="w-full mb-6">
                    {/* lucatodo: disable button if all projects in batch seen */}
                    <Button type="primary" full square href="/judge/live">
                        Next Project
                    </Button>
                    <div className="flex align-center justify-center mt-4">
                        <Button type="outline" square onClick={takeBreak} className="text-lg p-2">
                            I want to take a break!
                        </Button>
                    </div>
                </div>
                <div className="flex justify-evenly">
                    <StatBlock name="Seen" value={judge?.seen_projects.length as number}/>
                    <StatBlock name="Total Projects" value={projCount}/>
                </div>
                <h2 className="text-primary text-xl font-bold mt-4">Rank Projects</h2>
                {/* lucatodo: only update scores on submission? query this (only relevant if prioritisation implemented, issue #5 */}
                <p className="italic text-light">NB: Relative project order is tracked on the admin panel even before your submit a batch.</p>
                <div className="h-[1px] w-full bg-light my-2"></div>
                <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragStart={handleDragStart}
                    onDragEnd={handleDragEnd}
                >
                    <SortableContext items={projects} strategy={verticalListSortingStrategy}>
                        {projects.map((item) => (
                            <SortableItem key={item.id} item={item}/>
                        ))}
                    </SortableContext>
                    <DragOverlay>
                        {activeId ? (
                            <SortableItem item={projects.find((p) => p.id === activeId)}/>
                        ) : null}
                    </DragOverlay>
                </DndContext>
                <div className="w-full mt-4">
                    <div className="flex justify-center italic text-light text-center">
                        {/* lucatodo: update X with variable limit */}
                        {/* lucatodo: text updates if judging is ended manually to allow 'early' submission (see issue #4) */}
                        Please rank all your projects.<br/>
                        You can only submit rankings in batches of X projects.
                    </div>
                    <Button type="primary" full square href="" disabled={!allRanked}>
                        {/* lucatodo: add button functionality (inc. alert to confirm submission) */}
                        Submit Rankings
                        <p className="text-sm italic">And move onto next batch</p>
                    </Button>
                </div>
            </Container>
        </>
    );
};

export default Judge;
