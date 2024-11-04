import {JudgeSortField, ProjectSortField} from "./enums";
// todo: don't use these magic .d files - messes with imports

interface Project {
    id: string;
    name: string;
    guild: string;
    location: string;
    description: string;
    url: string;
    try_link: string;
    video_link: string;
    challenge_list: string[];
    seen: number;
    active: boolean;
    score: number;
    last_activity: number;
}

interface PublicProject {
    name: string;
    guild: string;
    location: string;
    description: string;
    url: string;
    try_link: string;
    video_link: string;
    challenge_list: string;
}

interface Judge {
    email: string | undefined;  // only returned for /api/judge - todo: should be a new separate type
    name: string | undefined;  // only returned for /api/judge
    id: string;
    keycloak_user_id: string;
    active: boolean;
    read_welcome: boolean;
    notes: string;
    current: string;
    seen: number;
    seen_projects: JudgedProject[];
    current_rankings: string[];
    past_rankings: string[][];
    last_activity: number;
}

interface JudgeWithKeycloak {
    judge: Judge;  // see above
    preferred_names: string | undefined;  // can just be unset by user
    first_names: string;
    last_names: string;
}

interface Stats {
    num_projects: number;
    hidden_projects: number;
    avg_project_seen: number;
    avg_judge_seen: number;
    num_judges: number;
}

type SortField = ProjectSortField | JudgeSortField;

interface SortState<T extends SortField> {
    field: T;
    ascending: boolean;
}

// TODO: Change this...
type VotePopupState = 'vote' | 'skip' | 'flag';

interface VotingProjectInfo {
    curr_name: string;
    curr_location: string;
    prev_name: string;
    prev_location: string;
}

interface YesNoResponse {
    yes_no: number;
}

interface JudgedProject {
    project_id: string;
    categories: { [name: string]: number };
    notes: string;
    name: string;
    guild: string;
    location: string;
    description: string;
}

type JudgedProjectWithUrl = {
    url: string;
} & JudgedProject;

type SortableJudgedProject = {
    id: number;
} & JudgedProject;

interface ClockState {
    time: number;
    running: boolean;
}

interface ProjectCount {
    count: number;
}

interface BatchRankingSize {
    brs: number;
}

interface Flag {
    id: string;
    judge_id: string;
    project_id: string;
    time: number;
    project_name: string;
    judge_name: string;
    project_location: string;
    reason: string;
}

interface Options {
    curr_table_num: number;
    clock: ClockState;
    judging_timer: number;
    categories: string[];
    batch_ranking_size: number;
    min_views: number;
}

interface FetchResponse<T> {
    status: number;
    error: string;
    data: T | null;
}

interface Timer {
    judging_timer: number;
}

interface NextJudgeProject {
    project_id: string;
}

interface ScoredItem {
    id: string;
    score: number;
}
