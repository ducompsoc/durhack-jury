import { useEffect, useState } from 'react';
import useAdminStore from '../../../store';
import HeaderEntry from './HeaderEntry';
import { JudgeSortField } from '../../../enums';
import JudgeRow from './JudgeRow';

const JudgesTable = () => {
    const unsortedJudges = useAdminStore((state) => state.judges);
    const fetchJudges = useAdminStore((state) => state.fetchJudges);
    const [judges, setJudges] = useState<Judge[]>([]);
    const [checked, setChecked] = useState<boolean[]>([]);
    const [sortState, setSortState] = useState<SortState<JudgeSortField>>({
        field: JudgeSortField.None,
        ascending: true,
    });

    const handleCheckedChange = (e: React.ChangeEvent<HTMLInputElement>, i: number) => {
        setChecked({
            ...checked,
            [i]: e.target.checked,
        });
    };

    const updateSort = (field: SortField) => {
        if (sortState.field === field) {
            // If sorted on same field and descending, reset sort state
            if (!sortState.ascending) {
                setSortState({
                    field: JudgeSortField.None,
                    ascending: true,
                });
                setJudges(unsortedJudges);
                return;
            }

            // Otherwise, sort descending
            setSortState({
                field,
                ascending: false,
            });
        } else {
            // If in different sorted state, sort ascending on new field
            setSortState({
                field: field as JudgeSortField,
                ascending: true,
            });
        }
    };

    // On load, fetch judges
    useEffect(() => {
        fetchJudges();
    }, [fetchJudges]);

    // When judges change, update judges and sort
    useEffect(() => {
        setChecked(Array(unsortedJudges.length).fill(false));

        let sortFunc = (a: Judge, b: Judge) => 0;
        const asc = sortState.ascending ? 1 : -1;
        switch (sortState.field) {
            case JudgeSortField.Name:
                sortFunc = (a, b) => a.name.localeCompare(b.name) * asc;
                break;
            case JudgeSortField.Email:
                sortFunc = (a, b) => a.email.localeCompare(b.email) * asc;
                break;
            case JudgeSortField.Seen:
                sortFunc = (a, b) => (a.seen - b.seen) * asc;
                break;
            case JudgeSortField.BatchesSubmitted:
                sortFunc = (a, b) => (a.past_rankings.length - b.past_rankings.length) * asc;
                break;
            case JudgeSortField.Updated:
                sortFunc = (a, b) => (a.last_activity - b.last_activity) * asc;
                break;
        }
        setJudges(unsortedJudges.sort(sortFunc));
    }, [unsortedJudges, sortState]);

    return (
        <div className="w-full px-8 pb-4">
            <table className="table-fixed w-full text-lg">
                <tbody>
                    <tr>
                        <th className="w-12"></th>
                        <HeaderEntry
                            name="Name"
                            updateSort={updateSort}
                            sortField={JudgeSortField.Name}
                            sortState={sortState}
                            align="left"
                        />
                        <HeaderEntry
                            name="Email"
                            updateSort={updateSort}
                            sortField={JudgeSortField.Email}
                            sortState={sortState}
                        />
                        <HeaderEntry
                            name="Seen"
                            updateSort={updateSort}
                            sortField={JudgeSortField.Seen}
                            sortState={sortState}
                        />
                        <HeaderEntry
                            name="Num. Batches Submitted"
                            updateSort={updateSort}
                            sortField={JudgeSortField.BatchesSubmitted}
                            sortState={sortState}
                        />
                        <HeaderEntry
                            name="Updated"
                            updateSort={updateSort}
                            sortField={JudgeSortField.Updated}
                            sortState={sortState}
                        />
                        <th className="text-right w-24">Actions</th>
                    </tr>
                    {judges.map((judge: Judge, idx) => (
                        <JudgeRow
                            key={idx}
                            idx={idx}
                            judge={judge}
                            checked={checked[idx]}
                            handleCheckedChange={handleCheckedChange}
                        />
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default JudgesTable;
