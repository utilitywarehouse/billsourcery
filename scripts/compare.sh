#!/bin/bash
TESTBRANCH=testing

for BRANCH in master ${TESTBRANCH} ; do
	echo BRANCH ${BRANCH}
	git checkout ${BRANCH} || exit -1
	go install || exit -1
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ calls-neo > neo_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ calls-dot > dot_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ executes > exec_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ methods > meth_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ reports > reports_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ public-procs > pp_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ forms > form_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ called-missing-methods > mm_${BRANCH}.txt
	billsourcery  --source-root /home/martin/uw_work/uw-bill-source-history/ all-modules > all_${BRANCH}.txt
done

for PREF in neo dot exec meth reports pp form mm all ; do
	echo ${PREF}
	for BRANCH in master ${TESTBRANCH} ; do
		sort ${PREF}_${BRANCH}.txt | b3sum
		#b3sum ${PREF}_${BRANCH}.txt
	done
	echo
done

