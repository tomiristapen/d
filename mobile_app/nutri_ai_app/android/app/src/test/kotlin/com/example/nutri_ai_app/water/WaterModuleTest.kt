package com.example.nutri_ai_app.water

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset

class WaterModuleTest {

    private val fixedClock: Clock =
        Clock.fixed(Instant.parse("2026-03-19T00:00:00Z"), ZoneOffset.UTC)

    @Test
    fun mapStepsToActivity_mapsBoundaries() {
        val activityService = ActivityService(googleFitService = null)

        assertEquals(ActivityLevel.LOW, activityService.mapStepsToActivity(0))
        assertEquals(ActivityLevel.LOW, activityService.mapStepsToActivity(4_999))
        assertEquals(ActivityLevel.MODERATE, activityService.mapStepsToActivity(5_000))
        assertEquals(ActivityLevel.MODERATE, activityService.mapStepsToActivity(10_000))
        assertEquals(ActivityLevel.HIGH, activityService.mapStepsToActivity(10_001))
    }

    @Test(expected = IllegalArgumentException::class)
    fun mapStepsToActivity_rejectsNegativeSteps() {
        val activityService = ActivityService(googleFitService = null)
        activityService.mapStepsToActivity(-1)
    }

    @Test
    fun calculateWaterGoal_usesActivityMultiplier_andSmallGoalAdjustment() {
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 70.0,
                        goal = Goal.WEIGHT_LOSS,
                        activitySource = ActivitySource.MANUAL,
                        manualActivity = ActivityLevel.MODERATE,
                    ),
                activityService = ActivityService(googleFitService = null),
                clock = fixedClock,
            )

        // MODERATE => 35 ml/kg => 70*35=2450; WEIGHT_LOSS => +200 => 2650
        val goalMl =
            waterService.calculateWaterGoal(
                weightKg = 70.0,
                activityLevel = ActivityLevel.MODERATE,
                goal = Goal.WEIGHT_LOSS,
            )
        assertEquals(2_650, goalMl)
    }

    @Test(expected = IllegalArgumentException::class)
    fun calculateWaterGoal_rejectsInvalidWeight() {
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 70.0,
                        goal = Goal.HEALTHY,
                        activitySource = ActivitySource.MANUAL,
                        manualActivity = ActivityLevel.LOW,
                    ),
                activityService = ActivityService(googleFitService = null),
                clock = fixedClock,
            )

        waterService.calculateWaterGoal(
            weightKg = 0.0,
            activityLevel = ActivityLevel.LOW,
            goal = Goal.HEALTHY,
        )
    }

    @Test
    fun addWater_accumulatesDailyTotal() {
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 70.0,
                        goal = Goal.MAINTAIN,
                        activitySource = ActivitySource.MANUAL,
                        manualActivity = ActivityLevel.LOW,
                    ),
                activityService = ActivityService(googleFitService = null),
                clock = fixedClock,
            )

        waterService.addWater(250)
        waterService.addWater(500)

        assertEquals(750, waterService.getDailyTotal())
    }

    @Test(expected = IllegalArgumentException::class)
    fun addWater_rejectsNonPositiveAmounts() {
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 70.0,
                        goal = Goal.MAINTAIN,
                        activitySource = ActivitySource.MANUAL,
                        manualActivity = ActivityLevel.LOW,
                    ),
                activityService = ActivityService(googleFitService = null),
                clock = fixedClock,
            )

        waterService.addWater(0)
    }

    @Test
    fun calculateProgress_isCappedAt100Percent() {
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 50.0,
                        goal = Goal.MAINTAIN,
                        activitySource = ActivitySource.MANUAL,
                        manualActivity = ActivityLevel.LOW, // 30 ml/kg => 1500
                    ),
                activityService = ActivityService(googleFitService = null),
                clock = fixedClock,
            )

        waterService.addWater(2_000)
        val progress = waterService.calculateProgress()

        assertEquals(1.0, progress, 0.0001)
    }

    @Test
    fun googleFitOverridesManualWhenPermissionAndStepsAvailable() {
        val fakeGoogleFit =
            FakeGoogleFitService(
                hasPermission = true,
                steps = 12_000,
            )
        val activityService = ActivityService(googleFitService = fakeGoogleFit)
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 70.0,
                        goal = Goal.HEALTHY,
                        activitySource = ActivitySource.GOOGLE_FIT,
                        manualActivity = ActivityLevel.LOW,
                    ),
                activityService = activityService,
                clock = fixedClock,
            )

        val resolved = activityService.resolveActivityLevel(waterService.userProfile)
        assertEquals(ActivityLevel.HIGH, resolved)
    }

    @Test
    fun googleFitFallsBackToManual_whenNoPermissionOrNoStepsData() {
        val noPermission =
            FakeGoogleFitService(
                hasPermission = false,
                steps = 12_000,
            )
        val noSteps =
            FakeGoogleFitService(
                hasPermission = true,
                steps = null,
            )

        val profile =
            UserProfile(
                weightKg = 70.0,
                goal = Goal.HEALTHY,
                activitySource = ActivitySource.GOOGLE_FIT,
                manualActivity = ActivityLevel.MODERATE,
            )

        val activityNoPermission = ActivityService(googleFitService = noPermission)
        assertEquals(ActivityLevel.MODERATE, activityNoPermission.resolveActivityLevel(profile))

        val activityNoSteps = ActivityService(googleFitService = noSteps)
        assertEquals(ActivityLevel.MODERATE, activityNoSteps.resolveActivityLevel(profile))
    }

    @Test
    fun calculateProgress_usesResolvedActivityLevel() {
        val fakeGoogleFit =
            FakeGoogleFitService(
                hasPermission = true,
                steps = 10_000, // MODERATE
            )
        val activityService = ActivityService(googleFitService = fakeGoogleFit)
        val waterService =
            WaterService(
                userProfile =
                    UserProfile(
                        weightKg = 70.0,
                        goal = Goal.MUSCLE_GAIN,
                        activitySource = ActivitySource.GOOGLE_FIT,
                        manualActivity = ActivityLevel.LOW,
                    ),
                activityService = activityService,
                clock = fixedClock,
            )

        // MODERATE => 70*35=2450; MUSCLE_GAIN => +200 => 2650
        waterService.addWater(1_325)
        val progress = waterService.calculateProgress()
        assertTrue(progress > 0.49 && progress < 0.51)
    }
}

private class FakeGoogleFitService(
    private val hasPermission: Boolean,
    private val steps: Long?,
) : GoogleFitService {
    override fun hasPermission(): Boolean = hasPermission

    override fun readDailySteps(): Long? = steps
}
