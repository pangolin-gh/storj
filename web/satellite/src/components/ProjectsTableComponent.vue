// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <v-card>
        <v-text-field
            v-model="search"
            label="Search"
            :prepend-inner-icon="Search"
            single-line
            variant="solo-filled"
            flat
            hide-details
            clearable
            density="comfortable"
            rounded="lg"
            class="mx-2 mt-2"
        />

        <v-data-table
            :sort-by="sortBy"
            :headers="headers"
            :items="items"
            items-per-page="10"
            items-per-page-text="Projects per page"
            :search="search"
            no-data-text="No results found"
            item-key="path"
        >
            <template #item.name="{ item }">
                <v-btn
                    v-if="item.role !== ProjectRole.Invited"
                    class="rounded-lg pl-1 pr-4 ml-n1 justify-start font-weight-bold"
                    variant="text"
                    height="40"
                    color="default"
                    block
                    @click="openProject(item)"
                >
                    <img src="../assets/icon-project-tonal.svg" alt="Project" class="mr-3">
                    {{ item.name }}
                </v-btn>
                <div v-else class="pl-1 pr-4 ml-n1 d-flex align-center justify-start font-weight-bold">
                    <img src="../assets/icon-project-tonal.svg" alt="Project" class="mr-3">
                    <span class="text-no-wrap">{{ item.name }}</span>
                </div>
            </template>

            <template #item.role="{ item }">
                <v-chip :color="PROJECT_ROLE_COLORS[item.role]" size="small" class="font-weight-bold">
                    {{ item.role }}
                </v-chip>
            </template>

            <template #item.createdAt="{ item }">
                <span class="text-no-wrap">
                    {{ Time.formattedDate(item.createdAt) }}
                </span>
            </template>

            <template #item.actions="{ item }">
                <div class="w-100 d-flex align-center justify-space-between">
                    <v-btn
                        v-if="item.role === ProjectRole.Invited"
                        color="primary"
                        size="small"
                        :disabled="decliningIds.has(item.id)"
                        @click="emit('joinClick', item)"
                    >
                        Join Project
                    </v-btn>
                    <v-btn v-else color="primary" size="small" rounded="md" @click="openProject(item)">Open Project</v-btn>

                    <v-btn
                        v-if="item.role === ProjectRole.Owner || item.role === ProjectRole.Invited"
                        class="ml-2"
                        icon
                        color="default"
                        variant="outlined"
                        size="small"
                        rounded="md"
                        density="comfortable"
                        :loading="decliningIds.has(item.id)"
                    >
                        <v-icon :icon="Ellipsis" size="18" />
                        <v-menu activator="parent" location="bottom" transition="scale-transition">
                            <v-list class="pa-1">
                                <template v-if="item.role === ProjectRole.Owner">
                                    <v-list-item link @click="() => onSettingsClick(item)">
                                        <template #prepend>
                                            <component :is="Settings" :size="18" />
                                        </template>
                                        <v-list-item-title class="text-body-2 ml-3">
                                            Project Settings
                                        </v-list-item-title>
                                    </v-list-item>
                                    <v-divider class="my-1" />
                                    <v-list-item link @click="emit('inviteClick', item)">
                                        <template #prepend>
                                            <component :is="UserPlus" :size="18" />
                                        </template>
                                        <v-list-item-title class="text-body-2 ml-3">
                                            Add Members
                                        </v-list-item-title>
                                    </v-list-item>
                                </template>
                                <v-list-item v-else link @click="declineInvitation(item)">
                                    <template #prepend>
                                        <component :is="Trash2" :size="18" />
                                    </template>
                                    <v-list-item-title class="text-body-2 ml-3">
                                        Decline
                                    </v-list-item-title>
                                </v-list-item>
                            </v-list>
                        </v-menu>
                    </v-btn>
                </div>
            </template>
        </v-data-table>
    </v-card>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import {
    VCard,
    VTextField,
    VListItem,
    VChip,
    VBtn,
    VMenu,
    VList,
    VIcon,
    VListItemTitle,
    VDivider,
    VDataTable,
} from 'vuetify/components';
import { Ellipsis, Search, UserPlus, Settings, Trash2 } from 'lucide-vue-next';

import { Time } from '@/utils/time';
import { ProjectItemModel, PROJECT_ROLE_COLORS, ProjectInvitationResponse } from '@/types/projects';
import { ProjectRole } from '@/types/projectMembers';
import { useProjectsStore } from '@/store/modules/projectsStore';
import { useAnalyticsStore } from '@/store/modules/analyticsStore';
import { AnalyticsErrorEventSource, AnalyticsEvent } from '@/utils/constants/analyticsEventNames';
import { useNotify } from '@/utils/hooks';
import { ROUTES } from '@/router';
import { useBucketsStore } from '@/store/modules/bucketsStore';
import { DataTableHeader, SortItem } from '@/types/common';

defineProps<{
    items: ProjectItemModel[],
}>();

const emit = defineEmits<{
    (event: 'joinClick', item: ProjectItemModel): void;
    (event: 'inviteClick', item: ProjectItemModel): void;
}>();

const search = ref<string>('');
const decliningIds = ref(new Set<string>());

const analyticsStore = useAnalyticsStore();
const bucketsStore = useBucketsStore();
const projectsStore = useProjectsStore();

const router = useRouter();
const notify = useNotify();

const sortBy: SortItem[] = [{ key: 'name', order: 'asc' }];
const headers: DataTableHeader[] = [
    { title: 'Project', key: 'name', align: 'start' },
    { title: 'Role', key: 'role' },
    { title: 'Members', key: 'memberCount' },
    { title: 'Storage', key: 'storageUsed' },
    { title: 'Download', key: 'bandwidthUsed' },
    { title: 'Date Added', key: 'createdAt' },
    { title: 'Actions', key: 'actions', sortable: false, width: '0' },
];

/**
 * Selects the project and navigates to the project dashboard.
 */
function openProject(item: ProjectItemModel): void {
    // There is no reason to clear s3 data if the user is navigating to the previously selected project.
    if (projectsStore.state.selectedProject.id !== item.id) bucketsStore.clearS3Data();

    projectsStore.selectProject(item.id);

    router.push({
        name: ROUTES.Dashboard.name,
        params: { id: projectsStore.state.selectedProject.urlId },
    });
    analyticsStore.eventTriggered(AnalyticsEvent.NAVIGATE_PROJECTS);
}

/**
 * Selects the project and navigates to the project's settings.
 */
function onSettingsClick(item: ProjectItemModel): void {
    projectsStore.selectProject(item.id);
    router.push({
        name: ROUTES.ProjectSettings.name,
        params: { id: projectsStore.state.selectedProject.urlId },
    });
}

/**
 * Declines the project invitation.
 */
async function declineInvitation(item: ProjectItemModel): Promise<void> {
    if (decliningIds.value.has(item.id)) return;
    decliningIds.value.add(item.id);

    try {
        await projectsStore.respondToInvitation(item.id, ProjectInvitationResponse.Decline);
        analyticsStore.eventTriggered(AnalyticsEvent.PROJECT_INVITATION_DECLINED);
    } catch (error) {
        error.message = `Failed to decline project invitation. ${error.message}`;
        notify.notifyError(error, AnalyticsErrorEventSource.PROJECT_INVITATION);
    }

    try {
        await projectsStore.getUserInvitations();
        await projectsStore.getProjects();
    } catch (error) {
        error.message = `Failed to reload projects and invitations list. ${error.message}`;
        notify.notifyError(error, AnalyticsErrorEventSource.PROJECT_INVITATION);
    }

    decliningIds.value.delete(item.id);
}
</script>
