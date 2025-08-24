package com.featurevoting.api

import com.featurevoting.models.*
import retrofit2.Response
import retrofit2.http.*

interface ApiService {
    
    @POST("auth/login")
    suspend fun login(@Body loginRequest: LoginRequest): Response<LoginResponse>
    
    @GET("features")
    suspend fun getFeatures(
        @Query("page") page: Int = 1,
        @Query("per_page") perPage: Int = 20
    ): Response<FeaturesResponse>
    
    @POST("features")
    suspend fun createFeature(@Body createFeatureRequest: CreateFeatureRequest): Response<CreateFeatureResponse>
    
    @POST("features/{id}/vote")
    suspend fun voteForFeature(@Path("id") featureId: Int): Response<VoteResponse>
    
    @DELETE("features/{id}/vote")
    suspend fun removeVoteFromFeature(@Path("id") featureId: Int): Response<VoteResponse>
    
    @GET("auth/profile")
    suspend fun getProfile(): Response<User>
}