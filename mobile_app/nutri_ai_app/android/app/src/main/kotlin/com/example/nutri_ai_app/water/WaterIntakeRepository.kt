package com.example.nutri_ai_app.water

import java.time.LocalDate

interface WaterIntakeRepository {
    fun add(date: LocalDate, amountMl: Int)

    fun getTotal(date: LocalDate): Int
}

class InMemoryWaterIntakeRepository : WaterIntakeRepository {
    private val totalsByDate = mutableMapOf<LocalDate, Int>()

    override fun add(date: LocalDate, amountMl: Int) {
        val current = totalsByDate[date] ?: 0
        totalsByDate[date] = current + amountMl
    }

    override fun getTotal(date: LocalDate): Int = totalsByDate[date] ?: 0
}

